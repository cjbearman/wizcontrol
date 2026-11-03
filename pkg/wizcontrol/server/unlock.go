package server

import (
	"net/http"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/common"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/lights"
)

func (s *ServerInstance) unlockHandler(w http.ResponseWriter, r *http.Request) {
	l := r.PathValue("light")
	logger := common.Logger.GetLogger("server").WithField(FieldRemoteAddr, r.RemoteAddr).WithField(FieldOperation, "unlock").WithField(FieldLightOrGroup, l)

	affectedLights, err := lights.GetLightOrGroup(l)
	if err != nil || len(affectedLights) == 0 {
		WriteError(w, http.StatusNotFound, "Light or group not found")
		logger.Warn("Light or group not found")
		return
	}

	for _, light := range affectedLights {
		light.Locker.UnlockNow(false)
	}
	WriteSuccess(w)
	logger.Info("Light(s) unlocked")
}
