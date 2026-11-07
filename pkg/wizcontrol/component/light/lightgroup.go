package light

import "fmt"

// LightGroups represents a collection of light groups keyed by their name.
type LightGroups map[string]*LightGroup

// LightGroup represents a group of lights.
type LightGroup struct {
	name   string
	lights []*Light
}

// String returns a string representation of the light group.
func (lg *LightGroup) String() string {
	return fmt.Sprintf("LightGroup{name: %s, lights: %d}", lg.name, len(lg.lights))
}

// Name returns the name of the light group.
func (lg *LightGroup) Name() string {
	return lg.name
}

// Lights returns the lights that are part of the light group.
func (lg *LightGroup) Lights() []*Light {
	return lg.lights
}
