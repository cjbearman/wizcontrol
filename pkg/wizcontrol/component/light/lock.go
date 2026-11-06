package light

import (
	"sync"
	"time"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/common"
)

var KickFunc func() // Set by daemon to prevent import loop

type LightLock struct {
	lockedUntil time.Time
	lock        sync.Mutex
	timer       *time.Timer
	lightName   string
	lockToColor string
}

func newLocker(lightName string) *LightLock {
	return &LightLock{
		lightName: lightName,
	}
}

func (l *LightLock) LockFor(d time.Duration, color string) {
	logger := common.Logger.GetLogger("lightlock").WithField("name", l.lightName)
	l.lock.Lock()
	defer l.lock.Unlock()
	l.lockedUntil = time.Now().In(common.Location).Add(d)
	l.lockToColor = color
	l.setTimer(d)
	logger.Debugf("Light locked for duration: %v, until %v, color: %s", d, l.lockedUntil, color)
}

func (l *LightLock) LockUntil(t time.Time, color string) {
	logger := common.Logger.GetLogger("lightlock").WithField("name", l.lightName)
	l.lock.Lock()
	defer l.lock.Unlock()
	l.lockedUntil = t
	l.lockToColor = color
	l.setTimer(time.Until(t))
	logger.Debugf("Light locked until: %v, color: %s", l.lockedUntil, color)
}

func (l *LightLock) UnlockNow(noKick bool) {
	logger := common.Logger.GetLogger("lightlock").WithField("name", l.lightName)
	l.lock.Lock()
	defer l.lock.Unlock()
	l.cancelTimer()
	l.lockedUntil = time.Time{}
	logger.Debug("Light lock manually unlocked, kicking daemon")
	if !noKick {
		if KickFunc == nil {
			logger.Error("KickFunc is nil, cannot kick daemon")
		} else {
			KickFunc()
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
	if KickFunc != nil {
		KickFunc()
	}
}

func (l *LightLock) IsLocked() (bool, string, time.Time) {
	l.lock.Lock()
	defer l.lock.Unlock()
	return time.Now().In(common.Location).Before(l.lockedUntil), l.lockToColor, l.lockedUntil
}
