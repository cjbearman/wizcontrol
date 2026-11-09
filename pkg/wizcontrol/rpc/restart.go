package rpc

import (
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/config"
)

// HandleRestartRequest processes a request to restart the daemon
func HandleRestartRequest(cnf *config.Config) []byte {
	close(cnf.RestartCh)
	return NewGeneralResponse(true, "")
}
