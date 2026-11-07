package scene

import (
	"fmt"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/component/color"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/component/light"
)

type Element struct {
	targets []*light.Light
	color   string
}

type Scene struct {
	name     string
	elements []*Element
	group    string
}

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
			if colors.GetByName(e.Color) == nil {
				return nil, nil, fmt.Errorf("scene %q references undefined color %q", s.Name, e.Color)
			}

			targets := []*light.Light{}
			for _, t := range e.Targets {

				light := lights.GetByName(t)
				if light != nil {
					targets = append(targets, light)
				} else {
					group := lightgroups.GetByName(t)
					if group != nil {
						targets = append(targets, group.Lights()...)
					} else {
						return nil, nil, fmt.Errorf("scene %q references undefined target %q, this is neither a light nor a light group", s.Name, t)
					}
				}
			}

			element := &Element{
				targets: targets,
				color:   e.Color,
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

func (s *Scene) GetName() string {
	return s.name
}

func (s *Scene) GetGroup() string {
	return s.group
}

func (s Scenes) GetByName(name string) *Scene {
	return s[name]
}

func (s *Scene) String() string {
	return fmt.Sprintf("Scene{name: %s, group: %s, elements: %d	}", s.name, s.group, len(s.elements))
}

func (e *Element) GetTargets() []*light.Light {
	return e.targets
}

func (e *Element) GetColor() string {
	return e.color
}

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

func (s *Scene) GetElements() []*Element {
	return s.elements
}
