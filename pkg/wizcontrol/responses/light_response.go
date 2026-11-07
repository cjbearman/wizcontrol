package responses

import (
	"time"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/component/color"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/component/light"
)

type LightAttributes struct {
	IP                  string     `json:"ip,omitempty"`
	MAC                 string     `json:"mac,omitempty"`
	On                  bool       `json:"on"`
	ProgrammedColorName string     `json:"programmed_color_name"`
	ProgrammedColor     string     `json:"programmed_color"`
	ActualColorName     string     `json:"actual_color_name"`
	ActualColor         string     `json:"actual_color"`
	AffectedByScene     bool       `json:"affected_by_scene"`
	ActiveScene         string     `json:"active_scene"`
	Locked              bool       `json:"locked"`
	LockedUntil         *time.Time `json:"locked_until,omitempty"`
}

type LightStatus struct {
	Available  bool            `json:"available"`
	Attributes LightAttributes `json:"attributes"`
}

func BuildLightStatus(light *light.Light) LightStatus {
	lpc := light.LastProgrammedColor()
	var programmedColorName, programmedColor string
	if lpc != nil {
		programmedColorName, programmedColor = lpc.Name(), lpc.Value()
	} else {
		programmedColorName, programmedColor = color.UnknownColor, color.UnknownColor
	}

	var actualColor, actualColorName string
	lac := light.LastActualColor()
	if lac != nil {
		actualColorName, actualColor = lac.Name(), lac.Value()
	} else {
		actualColorName, actualColor = color.UnknownColor, color.UnknownColor
	}

	activeScene := light.ActiveScene()
	var lockedUntil *time.Time
	locked, _, until := light.Locker().IsLocked()
	if locked {
		// Kind of a special case, light is locked to an explicit color
		// so although it has a programmed color, it's not *really* being driven by a scene
		activeScene = ""
		lockedUntil = &until
	}

	var ip string
	if light.Device() != nil {
		ip = light.Device().IP()
	}

	return LightStatus{
		Available: light.Available(),
		Attributes: LightAttributes{
			IP:                  ip,
			MAC:                 light.MAC(),
			On:                  light.On() && light.Available(),
			ProgrammedColorName: programmedColorName,
			ProgrammedColor:     programmedColor,
			ActualColorName:     actualColorName,
			ActualColor:         actualColor,
			AffectedByScene:     activeScene != "",
			ActiveScene:         activeScene,
			Locked:              locked,
			LockedUntil:         lockedUntil,
		},
	}
}
