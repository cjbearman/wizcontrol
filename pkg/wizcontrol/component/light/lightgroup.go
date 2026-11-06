package light

import "fmt"

type LightGroups map[string]*LightGroup
type LightGroup struct {
	name   string
	lights []*Light
}

func (lg *LightGroup) String() string {
	return fmt.Sprintf("LightGroup{name: %s, lights: %d}", lg.name, len(lg.lights))
}
