package rpc

import (
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/common"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/config"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/daemon"
)

func HandleResetRequest(cnf *config.Config) []byte {
	logger := common.Logger.GetLogger("rpc-reset")

	cnf.AdditionalScenes.ClearAdditionalScenes()
	for _, light := range cnf.Lights {
		light.Locker().UnlockNow(true)
	}
	daemon.Instance.Kick()
	logger.Info("All lights unlocked and daily overrides reset")
	return NewGeneralResponse(true, "")
}
