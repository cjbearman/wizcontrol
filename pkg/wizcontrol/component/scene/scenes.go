package scene

// Scenes represents a collection of scenes, keyed by scene name.
type Scenes map[string]*Scene

// GetSceneElementsBySceneName retrieves the elements of a scene by its name. It returns the elements and a boolean indicating if the scene was found.
func (s Scenes) GetSceneElementsBySceneName(name string) ([]*Element, bool) {
	scene, found := s[name]
	if !found {
		return nil, false
	}
	return scene.elements, true
}
