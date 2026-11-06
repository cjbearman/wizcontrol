package scenes

var sceneGroups = make(map[string][]string)

func GetSceneGroup(name string) ([]string, bool) {
	group, exists := sceneGroups[name]
	return group, exists
}
