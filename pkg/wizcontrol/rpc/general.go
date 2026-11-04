package rpc

import (
	"encoding/json"
	"fmt"

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
