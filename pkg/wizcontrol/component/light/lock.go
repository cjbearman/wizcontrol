package light

import (
	"sync"
	"time"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/common"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/component/color"
)

// LightLock is a component used to manage locked statuses of lights
type LightLock struct {
	lockedUntil time.Time
	lock        sync.Mutex
	timer       *time.Timer
	lightName   string
	lockToColor *color.Color
	kicker      func()
}

// newLocker creates a new LightLock for the specified light name.
func newLocker(lightName string) *LightLock {
	return &LightLock{
		lightName: lightName,
	}
}

// LockFor locks the light for the specified duration with the given color value.
func (l *LightLock) LockFor(d time.Duration, color *color.Color) {
	logger := common.Logger.GetLogger("lightlock").WithField("name", l.lightName)
	l.lock.Lock()
	defer l.lock.Unlock()
	l.lockedUntil = time.Now().In(common.Location).Add(d)
	l.lockToColor = color
	l.setTimer(d)
	logger.Debugf("Light locked for duration: %v, until %v, color: %s", d, l.lockedUntil, color.Value())
}

// LockUntil locks the light until the specified time with the given color value.
func (l *LightLock) LockUntil(t time.Time, color *color.Color) {
	logger := common.Logger.GetLogger("lightlock").WithField("name", l.lightName)
	l.lock.Lock()
	defer l.lock.Unlock()
	l.lockedUntil = t
	l.lockToColor = color
	l.setTimer(time.Until(t))
	logger.Debugf("Light locked until: %v, color: %s", l.lockedUntil, color.Value())
}

// UnlockNow unlocks the light immediately. If noKick is false, it will trigger the kicker.
func (l *LightLock) UnlockNow(noKick bool) {
	logger := common.Logger.GetLogger("lightlock").WithField("name", l.lightName)
	l.lock.Lock()
	defer l.lock.Unlock()
	l.cancelTimer()
	l.lockedUntil = time.Time{}
	logger.Debug("Light lock manually unlocked, kicking daemon")
	if !noKick {
		if l.kicker != nil {
			l.kicker()
		} else {
			logger.Warnf("Kicker not set in light lock for %s", l.lightName)
		}
	}
}

// cancelTimer stops and clears the current timer for the light lock.
func (l *LightLock) cancelTimer() {
	logger := common.Logger.GetLogger("lightlock").WithField("name", l.lightName)
	if l.timer != nil {
		logger.Debug("Cancelling light lock timer")
		l.timer.Stop()
		l.timer = nil
	}
}

// setTimer sets a timer for the light lock for the specified duration.
func (l *LightLock) setTimer(d time.Duration) {
	logger := common.Logger.GetLogger("lightlock").WithField("name", l.lightName)
	l.cancelTimer()
	logger.Debugf("Setting light lock timer for: %v", d)
	l.timer = time.AfterFunc(d, l.timerExpired)
}

// timerExpired is called when the light lock timer expires.
func (l *LightLock) timerExpired() {
	logger := common.Logger.GetLogger("lightlock").WithField("name", l.lightName)
	l.lock.Lock()
	defer l.lock.Unlock()
	logger.Debug("Light lock expired, kicking daemon, light is now unlocked")
	l.lockedUntil = time.Time{}
	if l.kicker != nil {
		l.kicker()
	} else {
		logger.Warnf("Kicker not set in light lock for %s", l.lightName)
	}
}

// IsLocked returns whether the light is currently locked, the color value it is locked to, and the time until it is unlocked.
func (l *LightLock) IsLocked() (locked bool, colorValue *color.Color, t time.Time) {
	l.lock.Lock()
	defer l.lock.Unlock()
	return time.Now().In(common.Location).Before(l.lockedUntil), l.lockToColor, l.lockedUntil
}

// SetKicker sets the function to be called when the light lock expires or is manually unlocked.
// The function of the kicker is to inform the daemon to re-evaluate light settings.
func (l *LightLock) SetKicker(kicker func()) {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.kicker = kicker
}
