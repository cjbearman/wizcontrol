package scenes

import (
	"fmt"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/colors"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/lights"
)

// CnfScene describes a named lighting scene.
type CnfScene struct {
	Targets   []string `yaml:"targets"`
	ColorName string   `yaml:"color"`
}

// CnfTarget is a single target mapping from light/group name to a color name.
type CnfTarget map[string]string

type TargetsAndColor struct {
	Targets   []string
	ColorName string
}

var Scenes = make(map[string][]TargetsAndColor)

func AddScene(name string, targets []CnfScene) error {
	// Scene must be distinct
	if _, existing := Scenes[name]; existing {
		return fmt.Errorf("scene %q already exists", name)
	}

	sc := []TargetsAndColor{}

	for _, targetGroup := range targets {
		t := TargetsAndColor{
			ColorName: targetGroup.ColorName,
		}

		if _, exists := colors.Colors[targetGroup.ColorName]; !exists {
			return fmt.Errorf("scene %q lists invalid color %q", name, targetGroup.ColorName)
		}

		for _, target := range targetGroup.Targets {
			if !lights.IsValidLightOrGroup(target) {
				return fmt.Errorf("scene %q lists invalid target %q", name, target)
			}
			t.Targets = append(t.Targets, target)
		}

		sc = append(sc, TargetsAndColor{
			Targets:   t.Targets,
			ColorName: t.ColorName,
		})
	}
	Scenes[name] = sc

	return nil
}

// ScenesExist will return true if all scenes exist
// or false and the name of the first missing scene
func ScenesExist(names ...string) (bool, string) {
	for _, name := range names {
		if _, exists := Scenes[name]; !exists {
			return false, name
		}
	}
	return true, ""
}
