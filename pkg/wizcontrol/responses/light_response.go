package responses

import (
	"fmt"
	"time"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/component/color"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/component/light"
)

type HAColor struct {
	R  int `json:"r"`
	G  int `json:"g"`
	B  int `json:"b"`
	CW int `json:"c"`
	WW int `json:"w"`
}

type HAEntity struct {
	Brightness int     `json:"brightness,omitempty"`
	Mode       string  `json:"color_mode,omitempty"`
	Temp       int     `json:"color_temp,omitempty"`
	Color      HAColor `json:"color,omitempty"`
	RGBWW      string  `json:"rgbww,omitempty"`
	Effect     string  `json:"effect,omitempty"` // TODO: Finish up the implementation of this
	State      string  `json:"state"`
}

// LightAttributes represents the attributes of a light, including its IP, MAC, color, scene, and lock status.
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
	LastRSSI            *int       `json:"rssi,omitempty"`
	AvgRSSI             *int       `json:"avg_rssi,omitempty"`
}

// LightStatus represents the current status of a light, including its availability and attributes.
type LightStatus struct {
	Available  bool            `json:"available"`
	Attributes LightAttributes `json:"attributes"`
	HA         *HAEntity       `json:"ha"`
}

// BuildLightStatus constructs a LightStatus response from the given light's current state.
// 'light' is the light whose status is being represented.
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

	state := "OFF"
	if light.On() {
		state = "ON"
	}
	ha := &HAEntity{
		State: state,
	}

	raw := light.GetLastRawColor()
	if raw != nil {
		ha.Mode = "rgbww"
		ha.Color = HAColor{
			R:  raw.R,
			G:  raw.G,
			B:  raw.B,
			CW: raw.CW,
			WW: raw.WW,
		}
		ha.RGBWW = fmt.Sprintf("%d,%d,%d,%d,%d", raw.R, raw.G, raw.B, raw.CW, raw.WW)
		if raw.Temp != 0 {
			ha.Mode = "color_temp"
			ha.Temp = raw.Temp
		} // ELSE TODO SCENE
		ha.Brightness = raw.BR
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
			LastRSSI:            light.LastRSSI(),
			AvgRSSI:             light.AverageRSSI(),
		},
		HA: ha,
	}
}
