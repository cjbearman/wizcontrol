package colors

import "github.com/cjbearman/wiz/pkg/wiz"

// Colors maps color names to color values
var Colors = make(map[string]string)

// Preload all wiz scenes as colors under the same name
func init() {
	for _, scene := range wiz.AllSceneNames() {
		Colors[scene] = scene
	}
}
