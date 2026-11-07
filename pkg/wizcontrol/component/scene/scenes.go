package scene

type Scenes map[string]*Scene

func (s Scenes) GetSceneElementsBySceneName(name string) ([]*Element, bool) {
	scene, found := s[name]
	if !found {
		return nil, false
	}
	return scene.elements, true
}
