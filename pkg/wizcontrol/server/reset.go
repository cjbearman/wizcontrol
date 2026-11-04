package server

import (
	"net/http"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/common"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/rpc"
)

func (s *ServerInstance) resetHandler(w http.ResponseWriter, r *http.Request) {
	logger := common.Logger.GetLogger("server").WithField(FieldRemoteAddr, r.RemoteAddr).WithField(FieldOperation, "reset")
	logger.Info("reset request received")
	resp := rpc.HandleResetRequest()
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}
