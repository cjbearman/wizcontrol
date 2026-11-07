package rpc

import (
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/config"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/daemon"
)

// HandlePauseRequest processes a request to pause the daemon by setting its mode to off.
func HandlePauseRequest(cnf *config.Config) []byte {
	daemon.Instance.SetMode(daemon.DaemonModePaused)
	return NewGeneralResponse(true, "")
}
