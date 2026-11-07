package color

import (
	"fmt"
	"strings"

	"github.com/cjbearman/wiz/pkg/wiz"
)

const UnknownColor = "unknown"
const Off = "off"

type Color struct {
	name, value string
}

func Build(cfg *[]ConfigYAML) (Colors, error) {
	colors := make(Colors)
	colors.prefillWizColors()
	for _, c := range *cfg {
		if _, found := colors[c.Name]; found {
			return nil, fmt.Errorf("color %q is defined more than once", c.Name)
		}

		for name, oc := range colors {
			if strings.EqualFold(c.Color, oc.Value()) {
				return nil, fmt.Errorf("color %q has same value as existing color %q", c.Name, name)
			}
		}

		colors[c.Name] = &Color{
			name:  c.Name,
			value: strings.ToLower(c.Color),
		}
	}
	return colors, nil
}

func (c Color) Name() string {
	return c.name
}

func (c Color) Value() string {
	return c.value
}

func (c *Color) String() string {
	return fmt.Sprintf("%s (%s)", c.name, c.value)
}

func (c *Color) SameColorAs(other *Color) bool {
	if other == nil {
		return false
	}
	return c.value == other.value
}

func (c *Color) GetWizSettings() (*wiz.Settings, error) {
	return wiz.SettingsFromString(c.value)
}
