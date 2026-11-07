package color

import (
	"fmt"
	"strings"

	"github.com/cjbearman/wiz/pkg/wiz"
)

type Colors map[string]*Color

func (c Colors) GetByName(name string) *Color {
	if col, found := c[name]; found {
		return col
	}
	return nil
}

func (c Colors) GetByValue(color string) *Color {

	// We prefer one of our defined colors, but will return one with unknown spec if needd

	for _, col := range c {
		if strings.EqualFold(col.Value(), color) {
			// One of our own
			return col
		}
	}
	return &Color{
		// Unknown
		name:  UnknownColor,
		value: color,
	}
}

func (c Colors) prefillWizColors() {
	c[Off] = &Color{
		name:  Off,
		value: Off,
	}
	for _, wizScene := range wiz.AllSceneNames() {
		name := fmt.Sprintf("wiz_%s", strings.ToLower(wizScene))
		c[name] = &Color{
			name:  name,
			value: wizScene,
		}
	}
}
