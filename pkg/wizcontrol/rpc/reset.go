package rpc

import (
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/common"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/daemon"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/lights"
)

func HandleResetRequest() []byte {
	logger := common.Logger.GetLogger("rpc-reset")

	daemon.Daemon.ClearAdditionalScenes()
	for _, light := range lights.Lights {
		light.Locker.UnlockNow(true)
	}
	daemon.Daemon.Kick()
	logger.Info("All lights unlocked and daily overrides reset")
	return NewGeneralResponse(true, "All lights unlocked and daily overrides reset")
}
