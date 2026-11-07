package server

import (
	"net/http"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/common"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/rpc"
)

// unpauseHandler handles the request to unpause and returns the appropriate HTTP response.
func (s *Server) unpauseHandler(w http.ResponseWriter, r *http.Request) {
	logger := common.Logger.GetLogger("server").WithField(FieldRemoteAddr, r.RemoteAddr)
	logger.Info("unpause request received")
	resp := rpc.HandleUnpauseRequest(s.cnf)
	w.WriteHeader(http.StatusOK)
	w.Write(resp) // nolint:errcheck
}
