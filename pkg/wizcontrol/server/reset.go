package server

import (
	"net/http"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/common"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/daemon"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/lights"
)

func (s *ServerInstance) resetHandler(w http.ResponseWriter, r *http.Request) {
	logger := common.Logger.GetLogger("server").WithField(FieldRemoteAddr, r.RemoteAddr).WithField(FieldOperation, "reset")

	daemon.Daemon.ClearAdditionalScenes()
	for _, light := range lights.Lights {
		light.Locker.UnlockNow(true)
	}
	daemon.Daemon.Kick()
	WriteSuccess(w)
	logger.Info("All lights unloicked and daily overrides reset")
}
