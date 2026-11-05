package server

import (
	"io"
	"net/http"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/common"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/rpc"
)

func (s *Server) unlockHandler(w http.ResponseWriter, r *http.Request) {

	logger := common.Logger.GetLogger("server").WithField(FieldRemoteAddr, r.RemoteAddr)
	logger.Info("Unlock request received")

	rawBody, _ := io.ReadAll(r.Body)

	resp, err := rpc.HandleUnlockRequest(rawBody)
	if err != nil {
		w.WriteHeader(err.Status())
	} else {
		w.WriteHeader(http.StatusOK)
	}
	w.Write(resp)
}
