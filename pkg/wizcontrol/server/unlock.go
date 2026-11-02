package server

import (
	"net/http"

	"github.com/cjbearman/wizcontrol2/pkg/wizcontrol/common"
	"github.com/cjbearman/wizcontrol2/pkg/wizcontrol/lights"
)

func (s *ServerInstance) unlockHandler(w http.ResponseWriter, r *http.Request) {
	nocache(w)
	l := r.PathValue("light")
	logger := common.Logger.GetLogger("server").WithField(FieldRemoteAddr, r.RemoteAddr).WithField(FieldOperation, "unlock").WithField(FieldLightOrGroup, l)

	affectedLights, err := lights.GetLightOrGroup(l)
	if err != nil || len(affectedLights) == 0 {
		http.Error(w, "Light or group not found", http.StatusNotFound)
		logger.Warn("Light or group not found")
		return
	}

	for _, light := range affectedLights {
		light.Locker.UnlockNow(false)
	}
	w.WriteHeader(http.StatusOK)
	logger.Info("Light(s) unlocked")
}
