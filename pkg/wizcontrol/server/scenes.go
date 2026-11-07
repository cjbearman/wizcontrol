package server

import (
	"net/http"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/common"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/rpc"
)

// scenesHandler handles the request for scenes and returns the appropriate HTTP response.
func (s *Server) scenesHandler(w http.ResponseWriter, r *http.Request) {
	logger := common.Logger.GetLogger("server").WithField(FieldRemoteAddr, r.RemoteAddr)
	logger.Info("scenes request received")
	resp := rpc.HandleScenesRequest(s.cnf)
	w.WriteHeader(http.StatusOK)
	w.Write(resp) // nolint:errcheck
}
