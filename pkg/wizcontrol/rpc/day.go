package rpc

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/common"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/daemon"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/scenes"
)

type DayRequest struct {
	Scenes []string `json:"scenes"`
}

type DayResponse struct {
}

func HandleDayRequest(req []byte) ([]byte, *RpcError) {
	logger := common.Logger.GetLogger("rpc-day")
	request := &DayRequest{}

	err := json.Unmarshal(req, request)
	if err != nil {
		return NewGeneralResponse(false, fmt.Sprintf("invalid request: %v", err)), NewRpcError(400, "invalid request: %v", err)
	}

	if len(request.Scenes) == 0 {
		logger.Info("clearing day scenes")
		daemon.Daemon.ClearAdditionalScenes()
		daemon.Daemon.Kick()
		return NewGeneralResponse(true, ""), nil
	} else {
		for _, scene := range request.Scenes {
			if scenes.Scenes[scene] == nil {
				return NewGeneralResponse(false, fmt.Sprintf("Scene %s not found", scene)), NewRpcError(400, "Scene %s not found", scene)
			}
		}
	}

	today := time.Now().In(common.Location)
	endOfDay := time.Date(today.Year(), today.Month(), today.Day(), 23, 59, 59, 0, common.Location)

	daemon.Daemon.SetAdditionalScenes(request.Scenes, endOfDay)
	logger.Infof("setting daily scenes: %v", request.Scenes)
	daemon.Daemon.Kick()
	return NewGeneralResponse(true, ""), nil
}
