package light

import (
	"fmt"
	"sync"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/component/color"

	"github.com/cjbearman/wiz/pkg/wiz"
)

// Light represents an individual wiz lightbulb
type Light struct {
	name, ip, discoveryMAC  string
	mac                     string
	groups                  []string
	lastProgrammedColor     *color.Color
	lastProgrammedColorLock sync.Mutex
	lastActualColor         *color.Color
	lastActualColorLock     sync.Mutex
	rssi                    int
	cum_rssi, cnt_rssi      uint64
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

// Build constructs the Lights and LightGroups from the given configuration YAML
func Build(cfg *[]ConfigYAML) (Lights, LightGroups, error) {
	lights := make(Lights)

	// Validate each individual light
	for _, l := range *cfg {

		// A light must not have the same name as another light
		if _, found := lights[l.Name]; found {
			return nil, nil, fmt.Errorf("light %q is defined more than once", l.Name)
		}

		// A light must not have the same IP as another light
		if other := lights.getLightByIP(l.IP); other != nil {
			return nil, nil, fmt.Errorf("light %q has same IP as light %q", l.Name, other.name)
		}

		// A light must not have the same discovery MAC as another light
		if other := lights.getLightsByDiscoveryMac(l.MAC); other != nil {
			return nil, nil, fmt.Errorf("light %q has same MAC as light %q", l.Name, other.name)
		}

		// A WIZ device is only created at this time if we have an IP for the device
		// If we do not then we will be doing discovery, and the discovery will create the wiz device
		// once we discover it
		var d *wiz.Device
		if l.IP != "" {
			d = wiz.MustDevice(l.IP) // Safe since IP has previously been checked
			d.SetName(l.Name)        // Important that our devices are named for our lights, so we can correlate async results back
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
			// A light group must not have the same name as a light
			if _, exists := lights[groupName]; exists {
				return nil, nil, fmt.Errorf("light group %q has the same name as a light", groupName)
			}

			if groups[groupName] == nil {
				groups[groupName] = &LightGroup{name: groupName}
			}
			groups[groupName].lights = append(groups[groupName].lights, light)
		}
	}

	return lights, groups, nil
}

// String returns a string representation of the light
func (l *Light) String() string {
	return fmt.Sprintf("Light{name: %s, ip: %s, mac: %s, groups: %v}", l.name, l.ip, l.discoveryMAC, l.groups)
}

// LastProgrammedColor returns the last programmed color for the light, if any.
func (l *Light) LastProgrammedColor() *color.Color {
	l.lastProgrammedColorLock.Lock()
	defer l.lastProgrammedColorLock.Unlock()
	return l.lastProgrammedColor
}

// SetLastProgrammedColor sets the last programmed color for the light.
func (l *Light) SetLastProgrammedColor(c *color.Color) {
	l.lastProgrammedColorLock.Lock()
	defer l.lastProgrammedColorLock.Unlock()
	l.lastProgrammedColor = c
}

// LastActualColor returns the last actual color for the light, if any.
func (l *Light) LastActualColor() *color.Color {
	l.lastActualColorLock.Lock()
	defer l.lastActualColorLock.Unlock()
	return l.lastActualColor
}

// SetLastActualColor sets the last actual color for the light.
func (l *Light) SetLastActualColor(c *color.Color) {
	l.lastActualColorLock.Lock()
	defer l.lastActualColorLock.Unlock()
	l.lastActualColor = c
}

func (l *Light) Locker() *LightLock {
	return l.locker
}

// ActiveScene returns the active scene for the light, or empty string if no scene is set
func (l *Light) ActiveScene() string {
	l.activeSceneLock.Lock()
	defer l.activeSceneLock.Unlock()
	return l.activeScene
}

// SetActiveScene sets the active scene for the light.
func (l *Light) SetActiveScene(scene string) {
	l.activeSceneLock.Lock()
	defer l.activeSceneLock.Unlock()
	l.activeScene = scene
}

// Available returns whether the light is available.
func (l *Light) Available() bool {
	l.availableLock.Lock()
	defer l.availableLock.Unlock()
	return l.available
}

// SetAvailable sets the availability of the light.
func (l *Light) SetAvailable(available bool) {
	l.availableLock.Lock()
	defer l.availableLock.Unlock()
	l.available = available
}

// On returns whether the light is on.
func (l *Light) On() bool {
	l.onLock.Lock()
	defer l.onLock.Unlock()
	return l.on
}

// SetOn sets the on state of the light.
func (l *Light) SetOn(on bool) {
	l.onLock.Lock()
	defer l.onLock.Unlock()
	l.on = on
}

// Device returns the underlying wiz.Device for the light.
// this may be nil for devices undertaking discovery
func (l *Light) Device() *wiz.Device {
	l.deviceLock.Lock()
	defer l.deviceLock.Unlock()
	return l.device
}

// SetDevice sets the underlying wiz.Device for the light.
func (l *Light) SetDevice(device *wiz.Device) {
	l.deviceLock.Lock()
	defer l.deviceLock.Unlock()
	l.device = device
	l.device.SetName(l.name)
}

// UseDiscovery returns whether the light is using discovery mode.
func (l *Light) UseDiscovery() bool {
	return l.discoveryMAC != ""
}

// DiscoveryMAC returns the discovery MAC address for the light, if any.
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

// SetMAC sets the (non-discovery) MAC address for the light.
func (l *Light) SetMAC(mac string) {
	l.mac = mac
}

func (l *Light) RecordRSSI(rssi int) {
	l.rssi = rssi
	l.cum_rssi += uint64(rssi)
	l.cnt_rssi++
}

func (l *Light) AverageRSSI() *int {
	if l.cnt_rssi == 0 {
		return nil
	}
	avg := int(float64(l.cum_rssi) / float64(l.cnt_rssi))
	return &avg
}
func (l *Light) LastRSSI() int {
	return l.rssi
}
