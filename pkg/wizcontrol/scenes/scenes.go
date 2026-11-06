package scenes

import (
	"fmt"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/colors"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/lights"
)

type Scene struct {
	SceneItems []CnfScene `yaml:"elements"`
	Group      string     `yaml:"group"`
}

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

var scenes = make(map[string][]TargetsAndColor)

func GetScenes() map[string][]TargetsAndColor {
	return scenes
}

func AddScene(name string, group string, cnf []CnfScene) error {
	// Scene must be distinct
	if _, existing := scenes[name]; existing {
		return fmt.Errorf("scene %q already exists", name)
	}

	sc := []TargetsAndColor{}

	for _, targetGroup := range cnf {
		t := TargetsAndColor{
			ColorName: targetGroup.ColorName,
		}

		if _, exists := colors.GetColors()[targetGroup.ColorName]; !exists {
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

		if group != "" {
			sceneGroups[group] = append(sceneGroups[group], name)
		}
	}

	scenes[name] = sc

	return nil
}

// ScenesExist will return true if all scenes exist
// or false and the name of the first missing scene
func ScenesExist(names ...string) (bool, string) {
	for _, name := range names {
		if _, exists := scenes[name]; !exists {
			return false, name
		}
	}
	return true, ""
}
