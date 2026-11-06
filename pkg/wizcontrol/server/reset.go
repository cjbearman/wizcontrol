package server

import (
	"net/http"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/common"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/rpc"
)

func (s *Server) resetHandler(w http.ResponseWriter, r *http.Request) {
	logger := common.Logger.GetLogger("server").WithField(FieldRemoteAddr, r.RemoteAddr)
	logger.Info("reset request received")
	resp := rpc.HandleResetRequest(s.cnf)
	w.WriteHeader(http.StatusOK)
	w.Write(resp) // nolint:errcheck
}
