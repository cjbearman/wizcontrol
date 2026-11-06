package scene

import (
	"fmt"
	"sync"
	"time"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/common"
)

type SceneUntil struct {
	Scene string    `json:"scene"`
	Until time.Time `json:"until"`
}

type AdditionalScenes struct {
	additional []SceneUntil
	lock       sync.RWMutex
}

func (a *AdditionalScenes) AddAdditionalScene(scene string, until time.Time) error {
	a.lock.Lock()
	defer a.lock.Unlock()
	a.clearExpiredScenes()
	for _, existing := range a.additional {
		if existing.Scene == scene {
			return fmt.Errorf("scene %q is already added", scene)
		}
	}

	a.additional = append(a.additional, SceneUntil{
		Scene: scene,
		Until: until,
	})
	return nil
}

func (a *AdditionalScenes) DeleteAdditionalScene(scene string) error {
	a.lock.Lock()
	defer a.lock.Unlock()
	a.clearExpiredScenes()
	revised := []SceneUntil{}

	found := false
	for _, existing := range a.additional {
		if existing.Scene != scene {
			revised = append(revised, existing)
		} else {
			found = true
		}
	}

	if !found {
		return fmt.Errorf("scene %q not found", scene)
	}
	a.additional = revised
	return nil
}

func (a *AdditionalScenes) clearExpiredScenes() {
	revised := []SceneUntil{}
	for _, existing := range a.additional {
		if existing.Until.After(time.Now().In(common.Location)) {
			revised = append(revised, existing)
		}
	}
	a.additional = revised
}

func (a *AdditionalScenes) GetAdditionalScenesWithUntils() []SceneUntil {
	a.lock.RLock()
	defer a.lock.RUnlock()
	a.clearExpiredScenes()
	return append([]SceneUntil{}, a.additional...)
}

func (a *AdditionalScenes) GetAdditionalScenes() []string {
	a.lock.RLock()
	defer a.lock.RUnlock()
	a.clearExpiredScenes()
	result := []string{}
	for _, existing := range a.additional {
		result = append(result, existing.Scene)
	}
	return result
}

func (a *AdditionalScenes) GetAdditionalScenesAndTimes() []SceneUntil {
	a.lock.RLock()
	defer a.lock.RUnlock()
	a.clearExpiredScenes()
	return append([]SceneUntil{}, a.additional...)
}

func (a *AdditionalScenes) ClearAdditionalScenes() {
	a.lock.Lock()
	defer a.lock.Unlock()
	a.additional = []SceneUntil{}
}
