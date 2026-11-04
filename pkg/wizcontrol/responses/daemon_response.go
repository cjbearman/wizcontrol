package responses

import (
	"time"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/scenes"
)

type DaemonAttributes struct {
	Scenes                string              `json:"scenes"`
	CalendarScenes        []string            `json:"calendar_scenes"`
	AdditionalScenes      []scenes.SceneUntil `json:"additional_scenes"`
	AdditionalScenesUntil *time.Time          `json:"additional_scenes_until,omitempty"`
}

type DaemonStatus struct {
	Mode       string           `json:"mode"`
	Attributes DaemonAttributes `json:"attributes"`
}

func BuildDaemonResponse(
	scenes string,
	calendarScenes []string,
	additionalScenes []scenes.SceneUntil,
	mode string,
) DaemonStatus {
	return DaemonStatus{
		Attributes: DaemonAttributes{
			Scenes:           scenes,
			CalendarScenes:   calendarScenes,
			AdditionalScenes: additionalScenes,
		},
		Mode: mode,
	}
}
