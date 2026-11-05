package rpc

import (
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/common"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/daemon"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/lights"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/scenes"
)

func HandleResetRequest() []byte {
	logger := common.Logger.GetLogger("rpc-reset")

	scenes.ClearAdditionalScenes()
	for _, light := range lights.GetLights() {
		light.Locker.UnlockNow(true)
	}
	daemon.Instance.Kick()
	logger.Info("All lights unlocked and daily overrides reset")
	return NewGeneralResponse(true, "")
}
