// Package responses contains JSON formatting for daemon and light RPC operations
// which may be obtained through HTTP or MQTT APIs.
package responses

import (
	"time"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/component/scene"
)

// DaemonAttributes represents the attributes of the daemon, including scenes and their schedules.
type DaemonAttributes struct {
	Scenes                string     `json:"scenes"`
	CalendarScenes        []string   `json:"calendar_scenes"`
	AdditionalScenes      []string   `json:"additional_scenes"`
	AdditionalScenesUntil *time.Time `json:"additional_scenes_until,omitempty"`
}

// DaemonStatus represents the overall status of the daemon, including its mode and attributes.
type DaemonStatus struct {
	Mode       string           `json:"mode"`
	Attributes DaemonAttributes `json:"attributes"`
}

// BuildDaemonResponse constructs a DaemonStatus response from the given scenes and mode.
// 'scenes' is the current active scenes, 'calendarScenes' are the scheduled scenes,
// 'additionalScenes' are any extra scenes with their active duration, and 'mode' is the daemon's mode.
func BuildDaemonResponse(
	scenes string,
	calendarScenes []string,
	additionalScenes []scene.SceneUntil,
	mode string,
) DaemonStatus {

	addScenes := []string{}
	for _, item := range additionalScenes {
		addScenes = append(addScenes, item.Scene.Name())
	}
	if calendarScenes == nil {
		// null prevention
		calendarScenes = []string{}
	}

	return DaemonStatus{
		Attributes: DaemonAttributes{
			Scenes:           scenes,
			CalendarScenes:   calendarScenes,
			AdditionalScenes: addScenes,
		},
		Mode: mode,
	}
}
