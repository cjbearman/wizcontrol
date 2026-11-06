package senders

import (
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/component/light"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/responses"
)

// We get import loops between mqtt, daemon and rpc if we try to have daemon and rpc access mqtt methods directly
// instead, have mqtt provide some functions here, when it's enabled,
// so daemon and rpc can import this without triggering import loops

var LightStatus func(light *light.Light) error
var DaemonStatus func(status responses.DaemonStatus) error

func SendLightStatus(light *light.Light) error {
	if LightStatus != nil {
		return LightStatus(light)
	}
	return nil
}

func SendDaemonStatus(status responses.DaemonStatus) error {
	if DaemonStatus != nil {
		return DaemonStatus(status)
	}
	return nil
}
