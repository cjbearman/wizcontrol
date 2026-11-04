package server

import (
	"io"
	"net/http"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/common"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/rpc"
)

func (s *ServerInstance) addScenesHandler(w http.ResponseWriter, r *http.Request) {
	logger := common.Logger.GetLogger("server").WithField(FieldRemoteAddr, r.RemoteAddr)
	logger.Info("Add scenes request received")

	rawBody, _ := io.ReadAll(r.Body)

	resp, err := rpc.HandleAddScenesRequest(rawBody)
	if err != nil {
		w.WriteHeader(err.Status())
	} else {
		w.WriteHeader(http.StatusOK)
	}
	w.Write(resp)
}
