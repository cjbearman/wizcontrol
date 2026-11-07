package scene

import (
	"fmt"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/component/color"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/component/light"
)

// Element represents an individual element within a scene, including its targets and color.
type Element struct {
	targets    []*light.Light
	targetsRaw []string
	color      *color.Color
}

// Scene represents a scene, including its name, elements, and group.
type Scene struct {
	name     string
	elements []*Element
	group    string
}

// Build constructs scenes and scene groups from the provided configuration, colors, lights, and light groups.
func Build(cfg *[]ConfigYAML, colors color.Colors, lights light.Lights, lightgroups light.LightGroups) (Scenes, SceneGroups, error) {
	scenes := make(Scenes)
	groups := make(SceneGroups)

	for _, s := range *cfg {
		if _, found := scenes[s.Name]; found {
			return nil, nil, fmt.Errorf("scene %q is defined more than once", s.Name)
		}
		scene := &Scene{
			name:  s.Name,
			group: s.Group,
		}

		for _, e := range s.Elements {
			color := colors[e.Color]
			if color == nil {
				return nil, nil, fmt.Errorf("scene %q references undefined color %q", s.Name, e.Color)
			}

			targets := []*light.Light{}
			for _, t := range e.Targets {

				light := lights[t]
				if light != nil {
					targets = append(targets, light)
				} else {
					group := lightgroups[t]
					if group != nil {
						targets = append(targets, group.Lights()...)
					} else {
						return nil, nil, fmt.Errorf("scene %q references undefined target %q, this is neither a light nor a light group", s.Name, t)
					}
				}
			}

			element := &Element{
				targets:    targets,
				color:      color,
				targetsRaw: e.Targets,
			}
			scene.elements = append(scene.elements, element)
		}
		scenes[s.Name] = scene
		if scene.group != "" {
			groups.add(scene.group, scene)
		}
	}
	return scenes, groups, nil
}

// Name returns the name of the scene.
func (s *Scene) Name() string {
	return s.name
}

// GetGroup returns the group of the scene.
// may be empty if no group is defined
func (s *Scene) GetGroup() string {
	return s.group
}

// String returns a string representation of the scene.
func (s *Scene) String() string {
	return fmt.Sprintf("Scene{name: %s, group: %s, elements: %d	}", s.name, s.group, len(s.elements))
}

// GetTargets returns the targets of the element.
func (e *Element) GetTargets() []*light.Light {
	return e.targets
}

// GetRawTargets gets the names of the lights and lights groups referenced by the element.
func (e *Element) GetRawTargets() []string {
	return e.targetsRaw
}

// GetColor returns the color of the element.
func (e *Element) GetColor() *color.Color {
	return e.color
}

// GetByGroup returns all scenes belonging to the specified group.
func (s Scenes) GetByGroup(group string) ([]*Scene, bool) {
	result := []*Scene{}
	for _, scene := range s {
		if scene.group == group {
			result = append(result, scene)
		}
	}
	if len(result) == 0 {
		return nil, false
	}
	return result, true
}

// GetElements returns the elements of the scene.
func (s *Scene) GetElements() []*Element {
	return s.elements
}
