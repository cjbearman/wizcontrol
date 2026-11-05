package state

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/cjbearman/wiz/pkg/wiz"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/colors"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/common"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/lights"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/scenes"
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

func WriteState() error {
	if StateRecoveryFile == "" {
		// Don't write state if no filename is provided
		return nil
	}

	// Compile the state
	state := State{
		LightLocks: []LockState{},
		SceneLocks: []SceneState{},
	}
	for name, light := range lights.GetLights() {
		locked, color, until := light.Locker.IsLocked()
		if locked {
			state.LightLocks = append(state.LightLocks, LockState{
				LightName: name,
				Color:     color,
				Until:     until,
			})
		}
	}

	for _, scene := range scenes.GetAdditionalScenesWithUntils() {
		state.SceneLocks = append(state.SceneLocks, SceneState{
			Scene: scene.Scene,
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

func RecoverState() error {
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
		if light, ok := lights.GetLights()[lock.LightName]; ok {

			if _, found := colors.GetColors()[lock.Color]; !found {
				// Not a defined color, maybe a custom color
				if _, err := wiz.SettingsFromString(lock.Color); err != nil {
					// Nope
					logger.Warnf("skipping unknown color %q for light %q", lock.Color, lock.LightName)
					continue
				}
			}

			light.Locker.LockUntil(lock.Until, lock.Color)
		}
	}

	for _, scene := range state.SceneLocks {
		if err := scenes.AddAdditionalScene(scene.Scene, scene.Until); err != nil {
			logger.Warnf("failed to add additional scene %q: %v", scene.Scene, err)
		}
	}

	logger.Info("state recovery completed")
	return nil
}
