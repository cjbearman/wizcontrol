package scenes

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

var additionalScenes []SceneUntil
var additionalScenesLock sync.RWMutex

func AddAdditionalScene(scene string, until time.Time) error {
	additionalScenesLock.Lock()
	defer additionalScenesLock.Unlock()
	clearExpiredScenes()
	for _, existing := range additionalScenes {
		if existing.Scene == scene {
			return fmt.Errorf("Scene %q is already added", scene)
		}
	}

	additionalScenes = append(additionalScenes, SceneUntil{
		Scene: scene,
		Until: until,
	})
	return nil
}

func DeleteAdditionalScene(scene string) error {
	additionalScenesLock.Lock()
	defer additionalScenesLock.Unlock()
	clearExpiredScenes()
	revised := []SceneUntil{}

	found := false
	for _, existing := range additionalScenes {
		if existing.Scene != scene {
			revised = append(revised, existing)
		} else {
			found = true
		}
	}

	if !found {
		return fmt.Errorf("Scene %q not found", scene)
	}
	additionalScenes = revised
	return nil
}

func clearExpiredScenes() {
	revised := []SceneUntil{}
	for _, existing := range additionalScenes {
		if existing.Until.After(time.Now().In(common.Location)) {
			revised = append(revised, existing)
		}
	}
	additionalScenes = revised
}

func GetAdditionalScenesWithUntils() []SceneUntil {
	additionalScenesLock.RLock()
	defer additionalScenesLock.RUnlock()
	clearExpiredScenes()
	result := []SceneUntil{}
	for _, existing := range additionalScenes {
		result = append(result, existing)
	}
	return result
}

func GetAdditionalScenes() []string {
	additionalScenesLock.RLock()
	defer additionalScenesLock.RUnlock()
	clearExpiredScenes()
	result := []string{}
	for _, existing := range additionalScenes {
		result = append(result, existing.Scene)
	}
	return result
}

func GetAdditionalScenesAndTimes() []SceneUntil {
	additionalScenesLock.RLock()
	defer additionalScenesLock.RUnlock()
	clearExpiredScenes()
	result := []SceneUntil{}
	for _, existing := range additionalScenes {
		result = append(result, existing)
	}
	return result
}

func ClearAdditionalScenes() {
	additionalScenesLock.Lock()
	defer additionalScenesLock.Unlock()
	additionalScenes = []SceneUntil{}
}
