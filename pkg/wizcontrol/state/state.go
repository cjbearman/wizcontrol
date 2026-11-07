// Package state handles state storage and recovery upon restart//
package state

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/common"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/config"
)

// LockState represents the lock state of a light, including its name, color, and expiration time.
type LockState struct {
	LightName string    `json:"light_name"`
	Color     string    `json:"color"`
	Until     time.Time `json:"until"`
}

// SceneState represents the lock state of a scene, including its name and expiration time.
type SceneState struct {
	Scene string    `json:"scene"`
	Until time.Time `json:"until"`
}

// State represents the overall state, including light locks and scene locks.
type State struct {
	LightLocks []LockState  `json:"light_locks"`
	SceneLocks []SceneState `json:"scene_locks"`
}

// WriteState writes the current state to the state recovery file.
func WriteState(cnf *config.Config) error {
	logger := common.Logger.GetLogger("state")
	if cnf.System.StateFile() == "" {
		// No state file configured, so we're not writing state
		return nil
	}

	// Compile the state
	state := State{
		LightLocks: []LockState{},
		SceneLocks: []SceneState{},
	}
	for name, light := range cnf.Lights {
		locked, color, until := light.Locker().IsLocked()
		if locked {
			state.LightLocks = append(state.LightLocks, LockState{
				LightName: name,
				Color:     color.Value(),
				Until:     until,
			})
		}
	}

	for _, scene := range cnf.AdditionalScenes.GetAdditionalScenesWithUntils() {
		state.SceneLocks = append(state.SceneLocks, SceneState{
			Scene: scene.Scene.Name(),
			Until: scene.Until,
		})
	}

	filename_tmp := cnf.System.StateFile() + ".tmp"

	fp, err := os.Create(filename_tmp)
	if err != nil {
		logger.Warnf("failed to create temporary state file %q: %v", filename_tmp, err)
		return fmt.Errorf("failed to create temporary state file %q: %v", filename_tmp, err)
	}

	err = json.NewEncoder(fp).Encode(state)
	_ = fp.Close()

	if err != nil {
		_ = os.Remove(filename_tmp)
		logger.Warnf("failed to encode state: %v", err)
		return fmt.Errorf("failed to encode state: %v", err)
	}

	if err := os.Rename(filename_tmp, cnf.System.StateFile()); err != nil {
		_ = os.Remove(filename_tmp)
		logger.Warnf("failed to rename temporary state file %q to %q: %v", filename_tmp, cnf.System.StateFile(), err)
		return fmt.Errorf("failed to rename temporary state file %q to %q: %v", filename_tmp, cnf.System.StateFile(), err)
	}

	return nil
}

// RecoverState recovers the state from the state recovery file and applies it to the current configuration.
func RecoverState(cnf *config.Config) error {
	logger := common.Logger.GetLogger("state-recovery")
	if cnf.System.StateFile() == "" {
		logger.Warn("State recovery not possible since state_file is not defined")
		return nil
	}
	fp, err := os.Open(cnf.System.StateFile())
	if err != nil {
		return fmt.Errorf("failed to open state file %q: %v", cnf.System.StateFile(), err)
	}
	defer fp.Close() // nolint:errcheck

	state := State{}
	if err := json.NewDecoder(fp).Decode(&state); err != nil {
		return fmt.Errorf("failed to decode state file %q: %v", cnf.System.StateFile(), err)
	}

	// Apply the recovered state
	for _, lock := range state.LightLocks {
		if light, ok := cnf.Lights[lock.LightName]; ok {
			light.Locker().LockUntil(lock.Until, cnf.Colors.GetByValue(lock.Color))
		}
	}

	for _, scene := range state.SceneLocks {

		sc := cnf.Scenes[scene.Scene]
		if sc == nil {
			logger.Warnf("Scene %q not found", scene.Scene)
			continue
		}

		if err := cnf.AdditionalScenes.AddAdditionalScene(sc, scene.Until); err != nil {
			logger.Warnf("failed to add additional scene %q: %v", scene.Scene, err)
		}
	}

	logger.Info("state recovery completed")
	return nil
}
