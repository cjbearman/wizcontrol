package server

import (
	"encoding/json"
	"net/http"

	"github.com/cjbearman/wizcontrol2/pkg/wizcontrol/common"
	"github.com/cjbearman/wizcontrol2/pkg/wizcontrol/daemon"
	"github.com/cjbearman/wizcontrol2/pkg/wizcontrol/lights"
)

type StatusResponse struct {
	SceneName string                 `json:"scene_name"`
	Mode      string                 `json:"mode"`
	Lights    map[string]LightStatus `json:"lights"`
}

type LightStatus struct {
	Color     string `json:"color"`
	Available bool   `json:"available"`
}

func (s *ServerInstance) statusHandler(w http.ResponseWriter, r *http.Request) {
	nocache(w)
	logger := common.Logger.GetLogger("server").WithField(FieldRemoteAddr, r.RemoteAddr).WithField(FieldOperation, "status")
	status := &StatusResponse{Lights: make(map[string]LightStatus)}
	status.Mode = string(daemon.Daemon.GetMode())
	status.SceneName = daemon.Daemon.GetActiveSceneNames()

	for _, light := range lights.Lights {
		programmedColorName, _ := light.GetLastProgrammedColor()
		status.Lights[light.Device.Name()] = LightStatus{
			Color:     programmedColorName,
			Available: light.Available(),
		}
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(status)
	logger.Info("Status sent")
}
