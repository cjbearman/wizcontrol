package responses

import (
	"time"
)

type DaemonStatus struct {
	Scenes                string     `json:"scenes"`
	CalendarScenes        []string   `json:"calendar_scenes"`
	AdditionalScenes      []string   `json:"additional_scenes"`
	AdditionalScenesUntil *time.Time `json:"additional_scenes_until,omitempty"`
	Mode                  string     `json:"mode"`
}

func BuildDaemonResponse(
	scenes string,
	calendarScenes []string,
	additionalScenes []string,
	additionalScenesUntil *time.Time,
	mode string,
) DaemonStatus {
	return DaemonStatus{
		Scenes:                scenes,
		CalendarScenes:        calendarScenes,
		AdditionalScenes:      additionalScenes,
		AdditionalScenesUntil: additionalScenesUntil,
		Mode:                  mode,
	}
}
