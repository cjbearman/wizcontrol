package colors

import (
	"strings"

	"github.com/cjbearman/wiz/pkg/wiz"
)

const UnknownColorName = "unknown"

// Colors maps color names to color values
var Colors = make(map[string]string)

// Preload all wiz scenes as colors under the same name
func init() {
	for _, scene := range wiz.AllSceneNames() {
		sceneLC := strings.ToLower(scene)
		Colors[sceneLC] = sceneLC
	}
}

func NameForColor(color string) string {
	for name, value := range Colors {
		if strings.EqualFold(value, color) {
			return name
		}
	}
	return UnknownColorName
}
