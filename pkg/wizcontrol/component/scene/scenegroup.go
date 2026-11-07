package scene

type SceneGroup []*Scene
type SceneGroups map[string][]*Scene

func (s SceneGroups) add(group string, scene *Scene) {
	s[group] = append(s[group], scene)
}
