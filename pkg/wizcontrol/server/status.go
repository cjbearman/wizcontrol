package server

import (
	"encoding/json"
	"net/http"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/colors"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/common"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/daemon"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/lights"
)

type StatusResponse struct {
	Scenes string                 `json:"scenes"`
	Mode   string                 `json:"mode"`
	Lights map[string]LightStatus `json:"lights"`
}

type LightStatus struct {
	ProgrammedColorName string `json:"programmed_color_name"`
	ProgrammedColor     string `json:"programmed_color"`
	ActualColorName     string `json:"actual_color_name"`
	ActualColor         string `json:"actual_color"`
	Available           bool   `json:"available"`
}

func (s *ServerInstance) statusHandler(w http.ResponseWriter, r *http.Request) {
	logger := common.Logger.GetLogger("server").WithField(FieldRemoteAddr, r.RemoteAddr).WithField(FieldOperation, "status")
	status := &StatusResponse{Lights: make(map[string]LightStatus)}
	status.Mode = string(daemon.Daemon.GetMode())
	status.Scenes = daemon.Daemon.GetActiveSceneNames()

	for _, light := range lights.Lights {
		programmedColorName, programmedColor := light.GetLastProgrammedColor()
		actualColor := light.LastColor()
		var actualColorName string = colors.UnknownColorName
		if actualColor == "" {
			actualColor = colors.UnknownColorName
		} else {
			actualColorName = colors.NameForColor(actualColor)
		}

		status.Lights[light.Device.Name()] = LightStatus{
			ProgrammedColorName: programmedColorName,
			ProgrammedColor:     programmedColor,
			ActualColorName:     actualColorName,
			ActualColor:         actualColor,
			Available:           light.Available(),
		}
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(status)
	logger.Info("Status sent")
}
