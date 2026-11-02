package daemon

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/cjbearman/framework/pkg/logger"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/calendar"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/common"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/lights"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/operations"
)

type DaemonMode string

const (
	DaemonModeOff  = "off"
	DaemonModeAuto = "auto"
)

var (
	Daemon           *DaemonInstance
	EvaluationPeriod time.Duration
	RecheckPeriod    time.Duration
)

// Daemon is the runner that handles automatic light setting based on configured calendar
type DaemonInstance struct {
	*common.Runner
	kicker           chan struct{}
	lastRecheck      time.Time
	mode             DaemonMode
	activeSceneNames string
	lock             sync.Mutex
	additionalScenes []string
	additionalUntil  time.Time
}

// NewDaemon returns a new daemon, not running
func NewDaemon() *DaemonInstance {
	d := &DaemonInstance{
		Runner:          common.NewRunner(),
		kicker:          make(chan struct{}, 1),
		lastRecheck:     time.Time{},
		additionalUntil: time.Time{},
		mode:            DaemonModeAuto,
	}
	d.Runner.Exec = d.run
	lights.KickFunc = d.Kick
	return d
}

func (d *DaemonInstance) GetActiveSceneNames() string {
	d.lock.Lock()
	defer d.lock.Unlock()
	return d.activeSceneNames
}

// Kick will send a kick on the kick channel, unless there is already one pending, in which case there's not much point
func (d *DaemonInstance) Kick() {
	select {
	case d.kicker <- struct{}{}:
	default:
	}
}

func (d *DaemonInstance) SetMode(mode DaemonMode) {
	d.lock.Lock()
	defer d.lock.Unlock()
	d.mode = mode
}

func (d *DaemonInstance) GetMode() DaemonMode {
	d.lock.Lock()
	defer d.lock.Unlock()
	return d.mode
}

func (d *DaemonInstance) SetAdditionalScenes(scenes []string, until time.Time) {
	d.lock.Lock()
	defer d.lock.Unlock()
	d.additionalScenes = scenes
	d.additionalUntil = until
}

func (d *DaemonInstance) ClearAdditionalScenes() {
	d.lock.Lock()
	defer d.lock.Unlock()
	d.additionalScenes = nil
	d.additionalUntil = time.Time{}
}

func (d *DaemonInstance) getAdditionalScenes() ([]string, time.Time) {
	d.lock.Lock()
	defer d.lock.Unlock()
	if time.Now().After(d.additionalUntil) {
		d.additionalScenes = nil
		d.additionalUntil = time.Time{}
	}
	return d.additionalScenes, d.additionalUntil
}

// Internal run method, called by Runner to execute the daemon loop
func (d *DaemonInstance) run() {

	// Make sure we mark the runner as closed when we exit
	defer close(d.Runner.Closed)

	if EvaluationPeriod <= 0 {
		panic("Invalid evaluation period")
	}
	if RecheckPeriod <= 0 {
		panic("Invalid recheck period")
	}

	logger := common.Logger.GetLogger("daemon").
		WithField("ev", fmt.Sprintf("%v", EvaluationPeriod)).
		WithField("re", fmt.Sprintf("%v", RecheckPeriod))

	logger.Info("starting")

	// Send a kick, this will ensure we run an immediate iteration on startup
	d.Kick()

	// Ticker will trigger subsequent runs
	ticker := time.NewTicker(EvaluationPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-d.kicker:
			// Triggered for immediate first run or on a "kick"
			// which occurs when the web server programs a new setting
			logger.Debug("Kicked")

			// We're going to do now, so we can reset our evaluation period for the next interval
			ticker.Reset(EvaluationPeriod)

			d.runIteration(logger)
		case <-ticker.C:
			// Triggered on our regular tick interval
			logger.Debug("Tick")
			d.runIteration(logger)
		case <-d.Runner.StopSignalled:
			// Triggered when the runner wants us to stop
			logger.Info("Stopped")
			return
		}
	}
}

func (d *DaemonInstance) runIteration(logger logger.Logger) {

	recheck := false
	if d.lastRecheck.Add(RecheckPeriod).Before(time.Now()) {
		d.lastRecheck = time.Now()
		recheck = true
	}
	logger = logger.WithField("recheck", recheck)
	logger.Info("daemon iteration")

	if d.GetMode() == DaemonModeOff {
		logger.Debug("Daemon is off, skipping iteration")
		return
	}

	activeScenes := calendar.GetCalendarScenes()
	activeSceneNames := strings.Join(activeScenes, ",")
	d.lock.Lock()
	d.activeSceneNames = activeSceneNames
	d.lock.Unlock()
	logger = logger.WithField("base-scenes", activeSceneNames)

	additionalScenes, _ := d.getAdditionalScenes()

	if len(additionalScenes) > 0 {
		logger = logger.WithField("add-scenes", strings.Join(additionalScenes, ","))
		activeScenes = append(activeScenes, additionalScenes...)
	}

	// Update light availability, and additionall request status update for found lights
	// unless we are within a recheck period, in which case we'll be doing status updates on all
	logger.Debugf("Update light availability")
	operations.UpdateAvailability(!recheck)

	if recheck {
		logger.Debugf("Recheck period reached, rechecking all light statuses")
		operations.UpdateStatusForAllAvailableLights(false)
	}

	logger.Debug("Apply scenes")
	operations.ApplyScenes(activeScenes, false)
}
