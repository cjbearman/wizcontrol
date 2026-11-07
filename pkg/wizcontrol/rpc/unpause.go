package rpc

import (
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/config"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/daemon"
)

// HandleUnpauseRequest processes a request to unpause the daemon by setting its mode to auto.
func HandleUnpauseRequest(cnf *config.Config) []byte {
	daemon.Instance.SetMode(daemon.DaemonModeAuto)
	return NewGeneralResponse(true, "")
}
