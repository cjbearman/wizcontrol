package server

import (
	"io"
	"net/http"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/common"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/rpc"
)

// lockHandler handles the request to lock and returns the appropriate HTTP response.
func (s *Server) lockHandler(w http.ResponseWriter, r *http.Request) {
	logger := common.Logger.GetLogger("server").WithField(FieldRemoteAddr, r.RemoteAddr)
	logger.Info("lock request received")
	rawBody, _ := io.ReadAll(r.Body)

	resp, err := rpc.HandleLockRequest(s.cnf, rawBody)
	if err != nil {
		logger.Errorf("Failed to handle lock request: %v", err)
		w.WriteHeader(err.Status())
		w.Write(resp) // nolint:errcheck
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resp) // nolint:errcheck
}
