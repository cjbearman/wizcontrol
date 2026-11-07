package scene

import (
	"fmt"
	"sync"
	"time"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/common"
)

// SceneUntil represents a scene that will be in effect until a specified time
type SceneUntil struct {
	Scene *Scene    `json:"scene"`
	Until time.Time `json:"until"`
}

// AdditionalScenes represents a list of additional scenes and target end times
type AdditionalScenes struct {
	additional []SceneUntil
	lock       sync.RWMutex
}

// AddAdditionalScene adds a new scene with a specified end time to the list of additional scenes. Returns an error if the scene is already added.
func (a *AdditionalScenes) AddAdditionalScene(scene *Scene, until time.Time) error {
	a.lock.Lock()
	defer a.lock.Unlock()
	a.clearExpiredScenes()
	for _, existing := range a.additional {
		if existing.Scene == scene {
			return fmt.Errorf("scene %q is already added", scene.Name())
		}
	}

	a.additional = append(a.additional, SceneUntil{
		Scene: scene,
		Until: until,
	})
	return nil
}

// GetAdditionalScenes returns a list of all additional scenes that have not yet expired.
func (a *AdditionalScenes) GetAdditionalScenes() []*Scene {
	a.lock.RLock()
	defer a.lock.RUnlock()
	a.clearExpiredScenes()
	result := []*Scene{}
	a.clearExpiredScenes()
	for _, existing := range a.additional {
		result = append(result, existing.Scene)
	}
	return result
}

// DeleteAdditionalScene removes a scene from the list of additional scenes. Returns an error if the scene is not found.
func (a *AdditionalScenes) DeleteAdditionalScene(scene *Scene) error {
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
		return fmt.Errorf("scene %q not found", scene.Name())
	}
	a.additional = revised
	return nil
}

// clearExpiredScenes removes any scenes that have passed their target end time.
func (a *AdditionalScenes) clearExpiredScenes() {
	revised := []SceneUntil{}
	for _, existing := range a.additional {
		if existing.Until.After(time.Now().In(common.Location)) {
			revised = append(revised, existing)
		}
	}
	a.additional = revised
}

// GetAdditionalScenesWithUntils returns a list of all additional scenes along with their target end times.
func (a *AdditionalScenes) GetAdditionalScenesWithUntils() []SceneUntil {
	a.lock.RLock()
	defer a.lock.RUnlock()
	a.clearExpiredScenes()
	return append([]SceneUntil{}, a.additional...)
}

// GetAdditionalSceneNames returns a list of the names of all additional scenes that have not yet expired.
func (a *AdditionalScenes) GetAdditionalSceneNames() []string {
	a.lock.RLock()
	defer a.lock.RUnlock()
	a.clearExpiredScenes()
	result := []string{}
	for _, existing := range a.additional {
		result = append(result, existing.Scene.Name())
	}
	return result
}

// GetAdditionalScenesAndTimes returns a list of all additional scenes along with their target end times.
func (a *AdditionalScenes) GetAdditionalScenesAndTimes() []SceneUntil {
	a.lock.RLock()
	defer a.lock.RUnlock()
	a.clearExpiredScenes()
	return append([]SceneUntil{}, a.additional...)
}

// ClearAdditionalScenes removes all additional scenes.
func (a *AdditionalScenes) ClearAdditionalScenes() {
	a.lock.Lock()
	defer a.lock.Unlock()
	a.additional = []SceneUntil{}
}
