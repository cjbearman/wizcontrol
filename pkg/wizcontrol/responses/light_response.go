package responses

import (
	"time"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/colors"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/lights"
)

type LightAttributes struct {
	ProgrammedColorName string     `json:"programmed_color_name"`
	ProgrammedColor     string     `json:"programmed_color"`
	ActualColorName     string     `json:"actual_color_name"`
	ActualColor         string     `json:"actual_color"`
	AffectedByScene     bool       `json:"affected_by_scene"`
	ActiveScene         string     `json:"active_scene"`
	Locked              bool       `json:"locked"`
	LockedUUntil        *time.Time `json:"locked_until,omitempty"`
}

type LightStatus struct {
	Available  bool            `json:"available"`
	Attributes LightAttributes `json:"attributes"`
}

func BuildLightStatus(light *lights.Light) LightStatus {
	programmedColorName, programmedColor := light.GetLastProgrammedColor()
	actualColor := light.LastColor()
	var actualColorName string = colors.UnknownColorName
	if actualColor == "" {
		actualColor = colors.UnknownColorName
	} else {
		actualColorName = colors.NameForColor(actualColor)
	}
	activeScene := light.ActiveScene()
	if activeScene == "" {
		programmedColorName = ""
		programmedColor = ""
	}

	var lockedUntil *time.Time
	locked, until := light.Locker.IsLocked()
	if locked {
		// Kind of a special case, light is locked to an explicit color
		// so although it has a programmed color, it's not *really* being driven by a scene
		activeScene = ""
		lockedUntil = &until
	}

	return LightStatus{
		Available: light.Available(),
		Attributes: LightAttributes{
			ProgrammedColorName: programmedColorName,
			ProgrammedColor:     programmedColor,
			ActualColorName:     actualColorName,
			ActualColor:         actualColor,
			AffectedByScene:     activeScene != "",
			ActiveScene:         activeScene,
			Locked:              locked,
			LockedUUntil:        lockedUntil,
		},
	}
}
