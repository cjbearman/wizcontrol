package server

import (
	"net/http"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/common"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/rpc"
)

func (s *Server) restartHandler(w http.ResponseWriter, r *http.Request) {
	logger := common.Logger.GetLogger("server").WithField(FieldRemoteAddr, r.RemoteAddr)
	logger.Info("restart request received")
	w.WriteHeader(http.StatusOK)
	w.Write(rpc.HandleRestartRequest(s.cnf))
}
