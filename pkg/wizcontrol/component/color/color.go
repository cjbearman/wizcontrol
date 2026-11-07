package color

import (
	"fmt"
	"strings"

	"github.com/cjbearman/wiz/pkg/wiz"
)

// UnknownColor is the name used for an unknown color
const UnknownColor = "unknown"

// Off is a special color used to represent a light that is turned off
const Off = "off"

// Color represents a named color with a color value
type Color struct {
	name, value string
}

// Build builds the colors from the already validated YAML configuration
// Returns the main colors list plus an additional list containing just the wiz colors
func Build(cfg *[]ConfigYAML, wizColorPrefix string) (Colors, Colors, error) {
	colors := make(Colors)
	wizColors := colors.prefillWizColors(wizColorPrefix)
	for _, c := range *cfg {
		if _, found := colors[c.Name]; found {
			return nil, nil, fmt.Errorf("color %q is defined more than once", c.Name)
		}

		for name, oc := range colors {
			if strings.EqualFold(c.Color, oc.Value()) {
				return nil, nil, fmt.Errorf("color %q has same value as existing color %q", c.Name, name)
			}
		}

		colors[c.Name] = &Color{
			name:  c.Name,
			value: strings.ToLower(c.Color),
		}
	}
	return colors, wizColors, nil
}

// Name returns the color name
func (c Color) Name() string {
	return c.name
}

// Value returns the color setting
func (c Color) Value() string {
	return c.value
}

// String returns a string representation of the color
func (c *Color) String() string {
	return fmt.Sprintf("%s (%s)", c.name, c.value)
}

// SameAsColor returns true if the color has the same value (but not necessarily the same name) as another color
func (c *Color) SameColorAs(other *Color) bool {
	if other == nil {
		return false
	}
	return c.value == other.value
}

// GetWizSettings returns the Wiz settings corresponding to the color value
func (c *Color) GetWizSettings() (*wiz.Settings, error) {
	return wiz.SettingsFromString(c.value)
}
