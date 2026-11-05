package server

import (
	"net/http"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/common"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/rpc"
)

func (s *Server) colorsHandler(w http.ResponseWriter, r *http.Request) {
	logger := common.Logger.GetLogger("server").WithField(FieldRemoteAddr, r.RemoteAddr)
	logger.Info("colors request received")
	resp := rpc.HandleColorsRequest()
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}
