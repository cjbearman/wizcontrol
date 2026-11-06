package light

import "fmt"

type Lights map[string]*Light

type Light struct {
	name, ip, mac string
	groups        []string
}

func Build(cfg *[]ConfigYAML) (Lights, LightGroups, error) {
	lights := make(Lights)
	// Have to check all lights for uniqueness
	for _, l := range *cfg {
		if _, found := lights[l.Name]; found {
			return nil, nil, fmt.Errorf("light %q is defined more than once", l.Name)
		}
		if other := lights.getLightByIP(l.IP); other != nil {
			return nil, nil, fmt.Errorf("light %q has same IP as light %q", l.Name, other.name)
		}
		if other := lights.getLightByMAC(l.MAC); other != nil {
			return nil, nil, fmt.Errorf("light %q has same MAC as light %q", l.Name, other.name)
		}

		lights[l.Name] = &Light{
			name:   l.Name,
			ip:     l.IP,
			mac:    l.MAC,
			groups: l.Groups,
		}
	}

	// Build lightgroups
	groups := make(LightGroups)
	for _, light := range lights {
		for _, groupName := range light.groups {
			group, exists := groups[groupName]
			if !exists {
				group = &LightGroup{name: groupName}
				groups[groupName] = group
			}
			group.lights = append(group.lights, light)
		}
	}

	return lights, groups, nil
}

func (l Lights) getLightByIP(ip string) *Light {
	for _, light := range l {
		if light.ip == ip {
			return light
		}
	}
	return nil
}

func (l Lights) getLightByMAC(mac string) *Light {
	for _, light := range l {
		if light.mac == "" {
			continue
		}
		if light.mac == mac {
			return light
		}
	}
	return nil
}

func (l *Light) String() string {
	return fmt.Sprintf("Light{name: %s, ip: %s, mac: %s, groups: %v}", l.name, l.ip, l.mac, l.groups)
}
