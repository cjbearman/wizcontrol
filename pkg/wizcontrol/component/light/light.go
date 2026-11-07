package light

import (
	"fmt"
	"sync"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/component/color"

	"github.com/cjbearman/wiz/pkg/wiz"
)

const UnknownColor = "unknown"

type Light struct {
	name, ip, discoveryMAC  string
	mac                     string
	groups                  []string
	lastProgrammedColor     *color.Color
	lastProgrammedColorLock sync.Mutex
	lastActualColor         *color.Color
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
		if other := lights.getLightsByDiscoveryMac(l.MAC); other != nil {
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
			discoveryMAC:        l.MAC,
			groups:              l.Groups,
			lastProgrammedColor: nil,
			lastActualColor:     nil,
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

func (l *Light) String() string {
	return fmt.Sprintf("Light{name: %s, ip: %s, mac: %s, groups: %v}", l.name, l.ip, l.discoveryMAC, l.groups)
}

func (l *Light) LastProgrammedColor() *color.Color {
	l.lastProgrammedColorLock.Lock()
	defer l.lastProgrammedColorLock.Unlock()
	return l.lastProgrammedColor
}

func (l *Light) SetLastProgrammedColor(c *color.Color) {
	l.lastProgrammedColorLock.Lock()
	defer l.lastProgrammedColorLock.Unlock()
	l.lastProgrammedColor = c
}

func (l *Light) LastActualColor() *color.Color {
	l.lastActualColorLock.Lock()
	defer l.lastActualColorLock.Unlock()
	return l.lastActualColor
}

func (l *Light) SetLastActualColor(c *color.Color) {
	l.lastActualColorLock.Lock()
	defer l.lastActualColorLock.Unlock()
	l.lastActualColor = c
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
	return l.discoveryMAC != ""
}

func (l *Light) DiscoveryMAC() string {
	return l.discoveryMAC
}

func (l *Light) Name() string {
	return l.name
}

// MAC returns the MAC address (if known) for this light, else empty
func (l *Light) MAC() string {
	if l.mac != "" {
		// Got a mac
		return l.mac
	}

	// Haven't got a mac, but if discovery MAC was set, we can use that
	return l.discoveryMAC
}

func (l *Light) SetMAC(mac string) {
	l.mac = mac
}
