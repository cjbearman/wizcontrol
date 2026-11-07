package state

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/common"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/config"
)

var StateRecoveryFile string

type LockState struct {
	LightName string    `json:"light_name"`
	Color     string    `json:"color"`
	Until     time.Time `json:"until"`
}

type SceneState struct {
	Scene string    `json:"scene"`
	Until time.Time `json:"until"`
}

type State struct {
	LightLocks []LockState  `json:"light_locks"`
	SceneLocks []SceneState `json:"scene_locks"`
}

func WriteState(cnf *config.Config) error {
	if StateRecoveryFile == "" {
		// Don't write state if no filename is provided
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
				Color:     color,
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

	filename_tmp := StateRecoveryFile + ".tmp"

	fp, err := os.Create(filename_tmp)
	if err != nil {
		return fmt.Errorf("failed to create temporary state file %q: %v", filename_tmp, err)
	}

	err = json.NewEncoder(fp).Encode(state)

	_ = fp.Close()
	if err != nil {
		_ = os.Remove(filename_tmp)
		return fmt.Errorf("failed to rename temporary state file %q to %q: %v", filename_tmp, StateRecoveryFile, err)
	}

	if err := os.Rename(filename_tmp, StateRecoveryFile); err != nil {
		_ = os.Remove(filename_tmp)
		return fmt.Errorf("failed to rename temporary state file %q to %q: %v", filename_tmp, StateRecoveryFile, err)
	}

	return nil
}

func RecoverState(cnf *config.Config) error {
	logger := common.Logger.GetLogger("state-recovery")
	if StateRecoveryFile == "" {
		logger.Info("state_recovery_file not set, state is not recovered")
		// Don't recover state if no filename is provided
		return nil
	}
	fp, err := os.Open(StateRecoveryFile)
	if err != nil {
		return fmt.Errorf("failed to open state file %q: %v", StateRecoveryFile, err)
	}
	defer fp.Close() // nolint:errcheck

	state := State{}
	if err := json.NewDecoder(fp).Decode(&state); err != nil {
		return fmt.Errorf("failed to decode state file %q: %v", StateRecoveryFile, err)
	}

	// Apply the recovered state
	for _, lock := range state.LightLocks {
		if light, ok := cnf.Lights[lock.LightName]; ok {
			light.Locker().LockUntil(lock.Until, lock.Color)
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
