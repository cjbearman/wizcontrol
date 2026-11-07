package color

import (
	"fmt"
	"strings"

	"github.com/cjbearman/wiz/pkg/wiz"
)

const DefaultWizColorPrefix = "wiz_"

// Colors is a map of Color names to Color objects
type Colors map[string]*Color

// GetByValue returns the color with the given value, or an unknown color if not found
func (c Colors) GetByValue(value string) *Color {

	// Search our existing colors for one that has this value
	for _, col := range c {
		if strings.EqualFold(col.Value(), value) {
			// One of our own
			return col
		}
	}

	// We didn't find a color, so return an "unknown" color for this value
	return &Color{
		// Unknown
		name:  UnknownColor,
		value: value,
	}
}

// prefillWizColors will create and map a color for all the wiz scene names, which
// are valid wiz color definitions
// It will also return a set of those colors for additional use cases
func (c Colors) prefillWizColors(prefix string) (wizColors Colors) {
	wizColors = make(Colors)
	c[Off] = &Color{
		name:  Off,
		value: Off,
	}
	for _, wizScene := range wiz.AllSceneNames() {
		name := fmt.Sprintf("%s%s", prefix, strings.ToLower(wizScene))
		col := &Color{
			name:  name,
			value: strings.ToLower(wizScene),
		}
		c[name], wizColors[name] = col, col
	}
	return wizColors
}
