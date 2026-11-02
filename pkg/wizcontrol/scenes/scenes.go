package scenes

import (
	"fmt"

	"github.com/cjbearman/wizcontrol2/pkg/wizcontrol/colors"
	"github.com/cjbearman/wizcontrol2/pkg/wizcontrol/lights"
)

// CnfScene describes a named lighting scene.
type CnfScene struct {
	Targets []CnfTarget `yaml:"targets"`
}

// CnfTarget is a single target mapping from light/group name to a color name.
type CnfTarget map[string]string

type TargetAndColor struct {
	Target    string
	ColorName string
}

var Scenes = make(map[string][]TargetAndColor)

func AddScene(name string, scene CnfScene) error {
	// Scene must be distinct
	if _, existing := Scenes[name]; existing {
		return fmt.Errorf("scene %q already exists", name)
	}

	targetsAndColors := []TargetAndColor{}

	for _, target := range scene.Targets {
		if len(target) != 1 {
			return fmt.Errorf("all targets in scene %q must have exactly one mapping from light/group to color", name)
		}
		var targetName, colorName string
		for k, v := range target {
			targetName = k
			colorName = v
			break
		}

		// Target must be a light or a light group
		if !lights.IsValidLightOrGroup(targetName) {
			return fmt.Errorf("target %q in scene %q is not a valid light or light group", targetName, name)
		}

		if _, exists := colors.Colors[colorName]; !exists {
			return fmt.Errorf("color %q in scene %q is not a valid color", colorName, name)
		}
		targetsAndColors = append(targetsAndColors, TargetAndColor{
			Target:    targetName,
			ColorName: colorName,
		})
	}
	Scenes[name] = targetsAndColors

	return nil
}
