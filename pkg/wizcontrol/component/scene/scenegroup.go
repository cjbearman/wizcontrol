package scene

// SceneGroup represents a group of scenes.
type SceneGroup []*Scene

// SceneGroups represents a collection of scene groups, keyed by group name.
type SceneGroups map[string][]*Scene

// add adds a scene to the specified group within the collection of scene groups.
func (s SceneGroups) add(group string, scene *Scene) {
	s[group] = append(s[group], scene)
}
