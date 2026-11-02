package lights

import (
	"sync"
	"time"

	"github.com/cjbearman/wizcontrol2/pkg/wizcontrol/common"
)

var KickFunc func() // Set by daemon to prevent import loop

type LightLock struct {
	lockedUntil time.Time
	lock        sync.Mutex
	timer       *time.Timer
	lightName   string
}

func newLocker(lightName string) *LightLock {
	return &LightLock{
		lightName: lightName,
	}
}

func (l *LightLock) LockFor(d time.Duration) {
	logger := common.Logger.GetLogger("lightlock").WithField("name", l.lightName)
	l.lock.Lock()
	defer l.lock.Unlock()
	l.lockedUntil = time.Now().In(common.Location).Add(d)
	l.setTimer(d)
	logger.Debugf("Light locked for duration: %v, until %v", d, l.lockedUntil)
}

func (l *LightLock) UnlockNow(noKick bool) {
	logger := common.Logger.GetLogger("lightlock").WithField("name", l.lightName)
	l.lock.Lock()
	defer l.lock.Unlock()
	l.cancelTimer()
	l.lockedUntil = time.Time{}
	logger.Debug("Light lock manually unlocked, kicking daemon")
	if !noKick && KickFunc != nil {
		KickFunc()
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

func (l *LightLock) IsLocked() bool {
	l.lock.Lock()
	defer l.lock.Unlock()
	return time.Now().In(common.Location).Before(l.lockedUntil)
}
