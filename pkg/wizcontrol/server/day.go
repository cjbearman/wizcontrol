package server

import (
	"io"
	"net/http"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/common"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/rpc"
)

type DayRequest struct {
	Scenes []string `json:"scenes"`
}

func (s *ServerInstance) dayHandler(w http.ResponseWriter, r *http.Request) {

	logger := common.Logger.GetLogger("server").WithField(FieldRemoteAddr, r.RemoteAddr)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Errorf("Failed to read request body: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	resp, err2 := rpc.HandleDayRequest(body)
	if err2 != nil {
		logger.Errorf("Failed to handle day request: %v", err)
		w.WriteHeader(err2.Status())
		w.Write(resp)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}
