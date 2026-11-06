package server

import (
	"encoding/json"
	"net/http"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/common"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/daemon"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/responses"
)

type StatusResponse struct {
	Server responses.DaemonStatus           `json:"server"`
	Lights map[string]responses.LightStatus `json:"lights"`
}

func (s *Server) statusHandler(w http.ResponseWriter, r *http.Request) {
	logger := common.Logger.GetLogger("server").WithField(FieldRemoteAddr, r.RemoteAddr)
	status := &StatusResponse{
		Lights: make(map[string]responses.LightStatus),
		Server: responses.BuildDaemonResponse(
			daemon.Instance.GetActiveSceneNames(),
			daemon.Instance.GetCalendarScenes(),
			s.cnf.AdditionalScenes.GetAdditionalScenesAndTimes(),
			string(daemon.Instance.GetMode()),
		),
	}

	for _, light := range s.cnf.Lights {
		status.Lights[light.Name()] = responses.BuildLightStatus(light)
	}

	_ = json.NewEncoder(w).Encode(status)
	logger.Info("Status sent")
}
