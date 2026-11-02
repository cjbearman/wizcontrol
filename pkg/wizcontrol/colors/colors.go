package colors

import (
	"strings"

	"github.com/cjbearman/wiz/pkg/wiz"
)

// Colors maps color names to color values
var Colors = make(map[string]string)

// Preload all wiz scenes as colors under the same name
func init() {
	for _, scene := range wiz.AllSceneNames() {
		sceneLC := strings.ToLower(scene)
		Colors["wiz_"+sceneLC] = sceneLC
	}
}
