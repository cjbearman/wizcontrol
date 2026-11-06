package daemon

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/cjbearman/logger"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/calendar"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/common"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/lights"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/operations"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/responses"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/scenes"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/senders"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/state"
)

type DaemonMode string

const (
	DaemonModeOff  = "off"
	DaemonModeAuto = "auto"
)

var (
	EvaluationPeriod time.Duration
	RecheckPeriod    time.Duration
)

var Instance *Daemon

// Daemon is the runner that handles automatic light setting based on configured calendar
type Daemon struct {
	*common.BaseRunner
	kicker           chan struct{}
	lastRecheck      time.Time
	mode             DaemonMode
	activeSceneNames string
	lock             sync.Mutex
	calendarScenes   []string
	additionalUntil  time.Time
}

// NewDaemon returns a new daemon, not running
func NewDaemon() *Daemon {
	d := &Daemon{
		BaseRunner:      common.NewRunner("daemon"),
		kicker:          make(chan struct{}, 1),
		lastRecheck:     time.Time{},
		additionalUntil: time.Time{},
		mode:            DaemonModeAuto,
	}
	d.Exec = d.run
	lights.KickFunc = d.Kick
	return d
}

func (d *Daemon) GetActiveSceneNames() string {
	d.lock.Lock()
	defer d.lock.Unlock()
	return d.activeSceneNames
}

func (d *Daemon) GetCalendarScenes() []string {
	d.lock.Lock()
	defer d.lock.Unlock()
	return d.calendarScenes
}

// Kick will send a kick on the kick channel, unless there is already one pending, in which case there's not much point
func (d *Daemon) Kick() {
	select {
	case d.kicker <- struct{}{}:
	default:
	}
}

func (d *Daemon) SetMode(mode DaemonMode) {
	d.lock.Lock()
	defer d.lock.Unlock()
	d.mode = mode
}

func (d *Daemon) GetMode() DaemonMode {
	d.lock.Lock()
	defer d.lock.Unlock()
	return d.mode
}

// Internal run method, called by Runner to execute the daemon loop
func (d *Daemon) run() {

	// Make sure we mark the runner as closed when we exit
	defer close(d.Closed)

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
		case <-d.StopSignalled:
			// Triggered when the runner wants us to stop
			logger.Info("Stopped")
			return
		}
	}
}

func (d *Daemon) runIteration(logger logger.Logger) {
	recheck := false
	if d.lastRecheck.Add(RecheckPeriod).Before(time.Now()) {
		d.lastRecheck = time.Now()
		recheck = true
	}
	logger = logger.WithField("recheck", recheck)
	logger.Info("daemon iteration")

	if d.GetMode() == DaemonModeOff {
		logger.Debug("Daemon is off, skipping iteration")
		_ = d.SendDaemonStatus()
		return
	}

	calendarScenes := calendar.GetCalendarScenes()
	additionalScenes := scenes.GetAdditionalScenes()
	activeScenes := append(calendarScenes, additionalScenes...)
	activeSceneNames := strings.Join(activeScenes, ",")

	d.lock.Lock()
	d.calendarScenes = calendarScenes
	d.activeSceneNames = activeSceneNames
	d.lock.Unlock()

	logger = logger.WithField("base-scenes", activeSceneNames).WithField("add-scenes", strings.Join(additionalScenes, ","))
	logger.Infof("Will apply the following scenes in order: %s", strings.Join(activeScenes, ","))
	// Update light availability, and additionall request status update for returning lights
	logger.Debugf("Update light availability")
	operations.UpdateAvailability()

	if recheck {
		logger.Debugf("Recheck period reached, rechecking all light statuses")
		operations.UpdateStatusForAllAvailableLights(false)
	}

	logger.Debugf("Apply scenes")
	operations.ApplyScenes(activeScenes, false, true)
	_ = d.SendDaemonStatus()
	_ = state.WriteState()
	logger.Info("daemon iteration completed")
}

func (d *Daemon) SendDaemonStatus() error {
	status := responses.BuildDaemonResponse(
		d.GetActiveSceneNames(),
		d.GetCalendarScenes(),
		scenes.GetAdditionalScenesAndTimes(),
		string(d.GetMode()),
	)

	_ = senders.SendDaemonStatus(status)

	return nil
}
