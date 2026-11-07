package server

import (
	"net/http"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/common"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/rpc"
)

// colorsHandler handles the colors request and returns the appropriate HTTP response.
func (s *Server) colorsHandler(w http.ResponseWriter, r *http.Request) {
	logger := common.Logger.GetLogger("server").WithField(FieldRemoteAddr, r.RemoteAddr)
	logger.Info("colors request received")
	resp := rpc.HandleColorsRequest(s.cnf)
	w.WriteHeader(http.StatusOK)
	w.Write(resp) // nolint:errcheck
}
