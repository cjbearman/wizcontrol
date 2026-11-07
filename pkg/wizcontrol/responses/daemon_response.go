package responses

import (
	"time"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/component/scene"
)

type DaemonAttributes struct {
	Scenes                string     `json:"scenes"`
	CalendarScenes        []string   `json:"calendar_scenes"`
	AdditionalScenes      []ScUntil  `json:"additional_scenes"`
	AdditionalScenesUntil *time.Time `json:"additional_scenes_until,omitempty"`
}

type ScUntil struct {
	Scene string    `json:"scene"`
	Until time.Time `json:"until"`
}

type DaemonStatus struct {
	Mode       string           `json:"mode"`
	Attributes DaemonAttributes `json:"attributes"`
}

func BuildDaemonResponse(
	scenes string,
	calendarScenes []string,
	additionalScenes []scene.SceneUntil,
	mode string,
) DaemonStatus {

	addscenes := []ScUntil{}
	for _, item := range additionalScenes {
		addscenes = append(addscenes, ScUntil{
			Scene: item.Scene.GetName(),
			Until: item.Until,
		})
	}

	return DaemonStatus{
		Attributes: DaemonAttributes{
			Scenes:           scenes,
			CalendarScenes:   calendarScenes,
			AdditionalScenes: addscenes,
		},
		Mode: mode,
	}
}
