package rpc

import "fmt"

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
