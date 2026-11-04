package server

import (
	"encoding/json"
	"net/http"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/common"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/daemon"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/lights"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/responses"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/scenes"
)

type StatusResponse struct {
	Server responses.DaemonStatus           `json:"server"`
	Lights map[string]responses.LightStatus `json:"lights"`
}

func (s *ServerInstance) statusHandler(w http.ResponseWriter, r *http.Request) {
	logger := common.Logger.GetLogger("server").WithField(FieldRemoteAddr, r.RemoteAddr)
	status := &StatusResponse{
		Lights: make(map[string]responses.LightStatus),
		Server: responses.BuildDaemonResponse(
			daemon.Daemon.GetActiveSceneNames(),
			daemon.Daemon.GetCalendarScenes(),
			scenes.GetAdditionalScenesAndTimes(),
			string(daemon.Daemon.GetMode()),
		),
	}

	for _, light := range lights.Lights {
		status.Lights[light.Device.Name()] = responses.BuildLightStatus(light)
	}

	_ = json.NewEncoder(w).Encode(status)
	logger.Info("Status sent")
}
