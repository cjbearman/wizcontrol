package server

import (
	"net/http"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/common"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/rpc"
)

// pauseHandler handles the request to pause and returns the appropriate HTTP response.
func (s *Server) pauseHandler(w http.ResponseWriter, r *http.Request) {
	logger := common.Logger.GetLogger("server").WithField(FieldRemoteAddr, r.RemoteAddr)
	logger.Info("pause request received")
	resp := rpc.HandlePauseRequest(s.cnf)
	w.WriteHeader(http.StatusOK)
	w.Write(resp) // nolint:errcheck
}
