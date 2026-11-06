package light

import (
	"fmt"
	"sync"

	"github.com/cjbearman/wiz/pkg/wiz"
)

const UnknownColor = "unknown"

type ColorNameAndColor struct {
	name  string
	color string
}

func (c ColorNameAndColor) Name() string {
	return c.name
}

func (c ColorNameAndColor) Value() string {
	return c.color
}

type Lights map[string]*Light

type Light struct {
	name, ip, mac           string
	groups                  []string
	lastProgrammedColor     ColorNameAndColor
	lastProgrammedColorLock sync.Mutex
	lastActualColor         ColorNameAndColor
	lastActualColorLock     sync.Mutex
	locker                  *LightLock
	activeScene             string
	activeSceneLock         sync.Mutex
	available               bool
	availableLock           sync.Mutex
	on                      bool
	onLock                  sync.Mutex
	device                  *wiz.Device
	deviceLock              sync.Mutex
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

		var d *wiz.Device
		if l.IP != "" {
			d = wiz.MustDevice(l.IP)
			d.SetName(l.Name) // Important that our devices are named for our lights, so we can correlate async results back
		}

		lights[l.Name] = &Light{
			name:                l.Name,
			ip:                  l.IP,
			mac:                 l.MAC,
			groups:              l.Groups,
			lastProgrammedColor: ColorNameAndColor{name: UnknownColor, color: UnknownColor},
			lastActualColor:     ColorNameAndColor{name: UnknownColor, color: UnknownColor},
			locker:              newLocker(l.Name),
			activeScene:         "",
			available:           false,
			on:                  false,
			device:              d,
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

func (l Lights) GetByName(name string) *Light {
	if light, found := l[name]; found {
		return light
	}
	return nil
}

func (l *Light) String() string {
	return fmt.Sprintf("Light{name: %s, ip: %s, mac: %s, groups: %v}", l.name, l.ip, l.mac, l.groups)
}

func (l *Light) LastProgrammedColor() ColorNameAndColor {
	l.lastProgrammedColorLock.Lock()
	defer l.lastProgrammedColorLock.Unlock()
	return l.lastProgrammedColor
}

func (l *Light) SetLastProgrammedColor(name, color string) {
	l.lastProgrammedColorLock.Lock()
	defer l.lastProgrammedColorLock.Unlock()
	l.lastProgrammedColor = ColorNameAndColor{name: name, color: color}
}

func (l *Light) LastActualColor() ColorNameAndColor {
	l.lastActualColorLock.Lock()
	defer l.lastActualColorLock.Unlock()
	return l.lastActualColor
}

func (l *Light) SetLastActualColor(name, color string) {
	l.lastActualColorLock.Lock()
	defer l.lastActualColorLock.Unlock()
	l.lastActualColor = ColorNameAndColor{name: name, color: color}
}

func (l *Light) Locker() *LightLock {
	return l.locker
}

func (l *Light) ActiveScene() string {
	l.activeSceneLock.Lock()
	defer l.activeSceneLock.Unlock()
	return l.activeScene
}

func (l *Light) SetActiveScene(scene string) {
	l.activeSceneLock.Lock()
	defer l.activeSceneLock.Unlock()
	l.activeScene = scene
}

func (l *Light) Available() bool {
	l.availableLock.Lock()
	defer l.availableLock.Unlock()
	return l.available
}

func (l *Light) SetAvailable(available bool) {
	l.availableLock.Lock()
	defer l.availableLock.Unlock()
	l.available = available
}

func (l *Light) On() bool {
	l.onLock.Lock()
	defer l.onLock.Unlock()
	return l.on
}

func (l *Light) SetOn(on bool) {
	l.onLock.Lock()
	defer l.onLock.Unlock()
	l.on = on
}

func (l *Light) Device() *wiz.Device {
	l.deviceLock.Lock()
	defer l.deviceLock.Unlock()
	return l.device
}

func (l *Light) SetDevice(device *wiz.Device) {
	l.deviceLock.Lock()
	defer l.deviceLock.Unlock()
	l.device = device
	l.device.SetName(l.name)
}

func (l *Light) UseDiscovery() bool {
	return l.mac != ""
}

func (l Lights) DiscoveryRequired() bool {
	for _, light := range l {
		if light.UseDiscovery() {
			return true
		}
	}
	return false
}

func (l *Light) MAC() string {
	return l.mac
}

func (l *Light) Name() string {
	return l.name
}

func GetLightOrGroup(lightOrGroup string, l Lights, g LightGroups) []*Light {
	for _, item := range l {
		if item.Name() == lightOrGroup {
			return []*Light{item}
		}
	}
	for _, group := range g {
		if group.Name() == lightOrGroup {
			return group.Lights()
		}
	}
	return nil
}

func (l Lights) GetAllLightsDeviceGroup() *wiz.DeviceGroup {
	dg := wiz.NewDeviceGroup()

	for _, light := range l {
		d := light.Device()
		if d != nil {
			dg.AddDevice(d)
		}
	}

	return dg
}
