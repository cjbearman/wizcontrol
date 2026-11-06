package color

import (
	"fmt"
	"strings"

	"github.com/cjbearman/wiz/pkg/wiz"
)

const UnknownColor = "unknown"
const Off = "off"

type Colors map[string]*Color

type Color struct {
	name, color string
}

func Build(cfg *[]ConfigYAML) (Colors, error) {
	colors := make(Colors)
	colors.prefillWizColors()
	for _, c := range *cfg {
		if _, found := colors[c.Name]; found {
			return nil, fmt.Errorf("color %q is defined more than once", c.Name)
		}
		if other := colors.GetByColor(c.Color); other != nil {
			return nil, fmt.Errorf("color %q has same value as color %q", c.Name, other.name)
		}
		colors[c.Name] = &Color{
			name:  c.Name,
			color: c.Color,
		}
	}
	return colors, nil
}

func (c Color) Name() string {
	return c.name
}

func (c Color) Value() string {
	return c.color
}

func (c Colors) GetByName(name string) *Color {
	if col, found := c[name]; found {
		return col
	}
	return nil
}

func (c Colors) GetByColor(color string) *Color {
	for _, col := range c {
		if col.color == color {
			return col
		}
	}
	return &Color{
		name:  UnknownColor,
		color: color,
	}
}

func (c *Color) String() string {
	return fmt.Sprintf("%s: %s", c.name, c.color)
}

func (c Colors) prefillWizColors() {
	c[Off] = &Color{
		name:  Off,
		color: Off,
	}
	for _, wizScene := range wiz.AllSceneNames() {
		name := fmt.Sprintf("wiz_%s", strings.ToLower(wizScene))
		c[name] = &Color{
			name:  name,
			color: wizScene,
		}
	}
}

func (c Colors) All() map[string]*Color {
	return c
}
