package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/common"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/daemon"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/scenes"
)

type DayRequest struct {
	Scenes []string `json:"scenes"`
}

func (s *ServerInstance) dayHandler(w http.ResponseWriter, r *http.Request) {
	nocache(w)
	logger := common.Logger.GetLogger("server").WithField(FieldRemoteAddr, r.RemoteAddr).WithField(FieldOperation, "set-day-scenes")
	req := &DayRequest{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid request body"))
		logger.Warnf("Invalid request body")
		return
	}

	if len(req.Scenes) == 0 {
		logger.Info("clearing day scenes")
	} else {
		for _, scene := range req.Scenes {
			if scenes.Scenes[scene] == nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write(fmt.Appendf([]byte{}, "Scene %s not found", scene))
				logger.Warnf("Requested scene not found")
				return
			}
		}
	}

	today := time.Now().In(common.Location)
	endOfDay := time.Date(today.Year(), today.Month(), today.Day(), 23, 59, 59, 0, common.Location)
	if len(req.Scenes) > 0 {
		daemon.Daemon.SetAdditionalScenes(req.Scenes, endOfDay)
	} else {
		daemon.Daemon.ClearAdditionalScenes()
	}

	daemon.Daemon.Kick()
	logger.WithField(FieldScene, strings.Join(req.Scenes, ", ")).Infof("Day scenes set by web request")
	w.WriteHeader(http.StatusOK)
}
