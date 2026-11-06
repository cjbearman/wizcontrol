package scene

import (
	"fmt"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/component/color"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/component/light"
)

type Scenes map[string]*Scene

type Element struct {
	targets []string
	color   string
}

type Scene struct {
	name     string
	elements []*Element
	group    string
}

func Build(cfg *[]ConfigYAML, colors color.Colors, lights light.Lights, lightgroups light.LightGroups) (Scenes, error) {
	scenes := make(Scenes)

	for _, s := range *cfg {
		if _, found := scenes[s.Name]; found {
			return nil, fmt.Errorf("scene %q is defined more than once", s.Name)
		}
		scene := &Scene{
			name:  s.Name,
			group: s.Group,
		}

		for _, e := range s.Elements {

			if colors.GetByName(e.Color) == nil {
				return nil, fmt.Errorf("scene %q references undefined color %q", s.Name, e.Color)
			}

			for _, t := range e.Targets {
				if lights.GetByName(t) == nil && lightgroups.GetByName(t) == nil {
					return nil, fmt.Errorf("scene %q references undefined target %q", s.Name, t)
				}
			}

			element := &Element{
				targets: e.Targets,
				color:   e.Color,
			}
			scene.elements = append(scene.elements, element)
		}
		scenes[s.Name] = scene
	}
	return scenes, nil

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
