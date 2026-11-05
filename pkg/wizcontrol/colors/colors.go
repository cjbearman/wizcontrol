package colors

import (
	"strings"

	"github.com/cjbearman/wiz/pkg/wiz"
)

const UnknownColorName = "unknown"
const ColorOff = "off"
const WizScenePrefix = "wiz_"

// Colors maps color names to color values
var colors = make(map[string]string)

func GetColors() map[string]string {
	return colors
}

// Preload all wiz scenes as colors under the same name
// and add "off" color
func init() {
	for _, scene := range wiz.AllSceneNames() {
		sceneLC := strings.ToLower(scene)
		colors[WizScenePrefix+sceneLC] = sceneLC
	}
	colors[ColorOff] = ColorOff
}

func NameForColor(color string) string {
	for name, value := range colors {
		if strings.EqualFold(value, color) {
			return name
		}
	}
	return UnknownColorName
}
