package responses

import (
	"time"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/component/light"
)

type LightAttributes struct {
	On                  bool       `json:"on"`
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

func BuildLightStatus(light *light.Light) LightStatus {
	lpc := light.LastProgrammedColor()
	programmedColorName := lpc.Name()
	programmedColor := lpc.Value()
	lac := light.LastActualColor()
	actualColor := lac.Value()
	actualColorName := lac.Name()
	activeScene := light.ActiveScene()
	var lockedUntil *time.Time
	locked, _, until := light.Locker().IsLocked()
	if locked {
		// Kind of a special case, light is locked to an explicit color
		// so although it has a programmed color, it's not *really* being driven by a scene
		activeScene = ""
		lockedUntil = &until
	}

	return LightStatus{
		Available: light.Available(),
		Attributes: LightAttributes{
			On:                  light.On() && light.Available(),
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
