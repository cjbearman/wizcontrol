package server

import (
	"io"
	"net/http"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/common"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/rpc"
)

type LightSetRequest struct {
	Color   string `json:"color"`
	LockFor int    `json:"for"`
	Force   bool   `json:"force"`
}

type LightRespItem struct {
	Success bool   `json:"success"`
	Reason  string `json:"reason,omitempty"`
}

func (s *ServerInstance) setHandler(w http.ResponseWriter, r *http.Request) {
	logger := common.Logger.GetLogger("server").WithField(FieldRemoteAddr, r.RemoteAddr)
	logger.Info("set request received")
	rawBody, _ := io.ReadAll(r.Body)

	resp, err := rpc.HandleSetRequest(rawBody)
	if err != nil {
		logger.Errorf("Failed to handle set request: %v", err)
		w.WriteHeader(err.Status())
		w.Write(resp)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}
