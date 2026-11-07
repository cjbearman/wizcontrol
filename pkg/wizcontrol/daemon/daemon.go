// Package daemon provides a service to provide periodic management of light state based on configured calendars and additional scenes.
package daemon

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/cjbearman/logger"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/common"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/component/bus"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/config"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/operations"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/responses"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/state"
)

// DaemonMode represents the mode in which the daemon operates.
type DaemonMode string

const (
	// DaemonModeOff indicates that the daemon is turned off and will not manage lights.
	DaemonModeOff = "off"

	// DaemonModeAuto indicates that the daemon will automatically manage lights based on configured calendars and additional scenes.
	DaemonModeAuto = "auto"
)

// Instance holds the global instance of the daemon. There should only ever be one daemon running.
var Instance *Daemon

// Daemon is the runner that handles automatic light setting based on configured calendar and additional scenes.
type Daemon struct {
	*common.BaseRunner
	kicker           chan struct{}
	lastRecheck      time.Time
	mode             DaemonMode
	activeSceneNames string
	lock             sync.Mutex
	calendarScenes   []string
	additionalUntil  time.Time
	evaluationPeriod time.Duration
	recheckPeriod    time.Duration
	cnf              *config.Config
}

// NewDaemon returns a new daemon, not running
func NewDaemon(cnf *config.Config) *Daemon {
	d := &Daemon{
		BaseRunner:       common.NewRunner("daemon"),
		kicker:           make(chan struct{}, 1),
		lastRecheck:      time.Time{},
		additionalUntil:  time.Time{},
		mode:             DaemonModeAuto,
		evaluationPeriod: cnf.System.EvaluationPeriod(),
		recheckPeriod:    cnf.System.RecheckPeriod(),
		cnf:              cnf,
	}
	d.Exec = d.run

	// Various bits and pieces need to send us a message that we need to run a cycle
	// However, direct references end up causing import loops so instead we set a function
	// that can be used to kick us in a few places...
	for _, l := range cnf.Lights {
		l.Locker().SetKicker(d.Kick)
	}

	// Also, for import loop prevention, we set the global instance here
	// not ideal, but there's only ever one daemon, so no real problem
	Instance = d

	return d
}

// GetActiveSceneNames returns a comma separated list of active scene names
func (d *Daemon) GetActiveSceneNames() string {
	d.lock.Lock()
	defer d.lock.Unlock()
	return d.activeSceneNames
}

// GetCalendarScenes returns a list of scenes that are currently active from the calendar
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

// SetMode sets the daemon run mode
func (d *Daemon) SetMode(mode DaemonMode) {
	d.lock.Lock()
	d.mode = mode
	d.lock.Unlock()
	d.SendDaemonStatus()
	d.Kick()
}

// GetMode returns the current daemon run mode
func (d *Daemon) GetMode() DaemonMode {
	d.lock.Lock()
	defer d.lock.Unlock()
	return d.mode
}

// Internal run method, called by Runner to execute the daemon loop
func (d *Daemon) run() {

	// Make sure we mark the runner as closed when we exit
	defer close(d.Closed)

	// A couple of sanity checks to ensure the evaluation and recheck periods are valid
	if d.evaluationPeriod <= 0 {
		panic("Invalid evaluation period")
	}
	if d.recheckPeriod <= 0 {
		panic("Invalid recheck period")
	}

	logger := common.Logger.GetLogger("daemon").
		WithField("ev", fmt.Sprintf("%v", d.evaluationPeriod)).
		WithField("re", fmt.Sprintf("%v", d.recheckPeriod))

	logger.Info("starting")

	// Send a kick, this will ensure we run an immediate iteration on startup
	d.Kick()

	// Ticker will trigger subsequent runs
	ticker := time.NewTicker(d.evaluationPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-d.kicker:
			// Triggered for immediate first run or on a "kick"
			// which occurs when the web server programs a new setting
			logger.Debug("Kicked")

			// We're going to do now, so we can reset our evaluation period for the next interval
			ticker.Reset(d.evaluationPeriod)

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

// runIteration runs a single iteration of the daemon loop
func (d *Daemon) runIteration(logger logger.Logger) {
	recheck := false
	if d.lastRecheck.Add(d.recheckPeriod).Before(time.Now()) {
		d.lastRecheck = time.Now()
		recheck = true
	}

	logger = logger.WithField("recheck", recheck)
	logger.Info("Evaluating scenes against configured settings")

	if d.GetMode() == DaemonModeOff {
		logger.Debug("Daemon is off, skipping iteration")
		_ = d.SendDaemonStatus()
		return
	}

	// calendarScenes are whichever scenes the calendar currently dictates should be active
	// there may be multiple and are ordered
	calendarScenes := d.cnf.Calendars.GetCalendarScenes(d.cnf.Scenes)

	// additionalScenes may have been provided by API, and are also ordered
	additionalScenes := d.cnf.AdditionalScenes.GetAdditionalScenes()

	// activeScenes are the calendar scenes, followed by additional scenes, also ordered
	activeScenes := append(calendarScenes, additionalScenes...)

	activeSceneNamesList := []string{}
	for _, s := range activeScenes {
		activeSceneNamesList = append(activeSceneNamesList, s.Name())
	}
	calendarSceneNamesList := []string{}
	for _, s := range calendarScenes {
		calendarSceneNamesList = append(calendarSceneNamesList, s.Name())
	}
	additionalSceneNamesList := []string{}
	for _, s := range additionalScenes {
		additionalSceneNamesList = append(additionalSceneNamesList, s.Name())
	}

	activeSceneNames := strings.Join(activeSceneNamesList, ",")

	// Update the daemon's internal state with the current calendar and active scenes
	d.lock.Lock()
	d.calendarScenes = calendarSceneNamesList
	d.activeSceneNames = activeSceneNames
	d.lock.Unlock()

	logger = logger.WithField("base-scenes", activeSceneNames).WithField("add-scenes", strings.Join(additionalSceneNamesList, ","))
	logger.Infof("Will apply the following scenes in order: %s", strings.Join(activeSceneNamesList, ","))

	// Update light availability, and additionall request status update for returning lights
	logger.Debugf("Update light availability")
	operations.UpdateAvailability(d.cnf)

	// Periodically, we will unconditionally get status updates for all lights
	if recheck {
		logger.Debugf("Recheck period reached, rechecking all light statuses")
		operations.UpdateStatusForAllAvailableLights(d.cnf, false)
	}

	// Apply the active scenes to the lights
	operations.ApplyScenes(d.cnf, activeScenes)

	// Send an updated daemon status to MQTT
	_ = d.SendDaemonStatus()

	// Update persistent state, if we are doing that
	_ = state.WriteState(d.cnf)

	logger.Info("Completed evaluation of scenes against configured settings")
}

// SendDaemonStatus dispatches current daemon status to MQTT, if running
func (d *Daemon) SendDaemonStatus() error {
	status := responses.BuildDaemonResponse(
		d.GetActiveSceneNames(),
		d.GetCalendarScenes(),
		d.cnf.AdditionalScenes.GetAdditionalScenesAndTimes(),
		string(d.GetMode()),
	)

	statusEncoded, err := json.Marshal(status)
	if err != nil {
		return fmt.Errorf("failed to encode daemon status: %v", err)
	}

	d.cnf.Bus.Publish(bus.Message{
		Topic:   "server",
		Payload: statusEncoded,
	})

	return nil
}
