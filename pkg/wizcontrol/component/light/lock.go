package light

import (
	"sync"
	"time"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/common"
)

type LightLock struct {
	lockedUntil time.Time
	lock        sync.Mutex
	timer       *time.Timer
	lightName   string
	lockToColor string
	kicker      func()
}

func newLocker(lightName string) *LightLock {
	return &LightLock{
		lightName: lightName,
	}
}

func (l *LightLock) LockFor(d time.Duration, colorValue string) {
	logger := common.Logger.GetLogger("lightlock").WithField("name", l.lightName)
	l.lock.Lock()
	defer l.lock.Unlock()
	l.lockedUntil = time.Now().In(common.Location).Add(d)
	l.lockToColor = colorValue
	l.setTimer(d)
	logger.Debugf("Light locked for duration: %v, until %v, color: %s", d, l.lockedUntil, colorValue)
}

func (l *LightLock) LockUntil(t time.Time, colorValue string) {
	logger := common.Logger.GetLogger("lightlock").WithField("name", l.lightName)
	l.lock.Lock()
	defer l.lock.Unlock()
	l.lockedUntil = t
	l.lockToColor = colorValue
	l.setTimer(time.Until(t))
	logger.Debugf("Light locked until: %v, color: %s", l.lockedUntil, colorValue)
}

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

func (l *LightLock) cancelTimer() {
	logger := common.Logger.GetLogger("lightlock").WithField("name", l.lightName)
	if l.timer != nil {
		logger.Debug("Cancelling light lock timer")
		l.timer.Stop()
		l.timer = nil
	}
}

func (l *LightLock) setTimer(d time.Duration) {
	logger := common.Logger.GetLogger("lightlock").WithField("name", l.lightName)
	l.cancelTimer()
	logger.Debugf("Setting light lock timer for: %v", d)
	l.timer = time.AfterFunc(d, l.timerExpired)
}

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

func (l *LightLock) IsLocked() (locked bool, colorValue string, t time.Time) {
	l.lock.Lock()
	defer l.lock.Unlock()
	return time.Now().In(common.Location).Before(l.lockedUntil), l.lockToColor, l.lockedUntil
}

func (l *LightLock) SetKicker(kicker func()) {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.kicker = kicker
}
