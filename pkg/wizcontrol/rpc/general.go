package rpc

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/common"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/lights"
)

// Will be set by MQTT (prevents import loop)
var LightStatusDispatcher func(*lights.Light) error

type GeneralResponse struct {
	Success bool    `json:"success"`
	Error   *string `json:"error,omitempty"`
}

func NewGeneralResponse(success bool, errmsg string) []byte {
	resp := GeneralResponse{
		Success: success,
	}
	if errmsg != "" {
		resp.Error = &errmsg
	}
	data, _ := json.Marshal(resp)
	return data
}

type RpcError struct {
	error  string
	status int
}

func (e *RpcError) Error() string {
	return e.error
}

func (e *RpcError) Status() int {
	return e.status
}

func NewRpcError(status int, format string, args ...interface{}) *RpcError {
	return &RpcError{
		error:  fmt.Sprintf(format, args...),
		status: status,
	}
}

func validateForUntil(forspec, until string) (time.Time, *RpcError) {
	if forspec == "" && until == "" {
		return time.Time{}, NewRpcError(400, "either 'for' or 'until' is required")
	}
	if forspec != "" && until != "" {
		return time.Time{}, NewRpcError(400, "'for' and 'until' are mutually exclusive")
	}
	if forspec != "" {
		duration, err := time.ParseDuration(forspec)
		if err != nil {
			return time.Time{}, NewRpcError(400, "invalid 'for' duration: %v", err)
		}
		return time.Now().In(common.Location).Add(duration), nil
	}
	if until != "" {
		// TODO: Support additional time formats
		lockUntil, err := time.ParseInLocation("2006-01-02 15:04", until, common.Location)
		if err != nil {
			return time.Time{}, NewRpcError(400, "invalid 'until' time: %v", err)
		}
		return lockUntil, nil
	}
	return time.Time{}, NewRpcError(400, "either 'for' or 'until' is required")
}
