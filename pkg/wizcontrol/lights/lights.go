package lights

import (
	"errors"
	"fmt"
	"net"
	"regexp"
	"strings"
	"sync"

	"github.com/cjbearman/logger"
	"github.com/cjbearman/wiz/pkg/wiz"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/common"
)

var macReg = regexp.MustCompile(`^([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})$`)

type Light struct {
	name                    string
	mac                     string
	ip                      string
	device                  *wiz.Device
	deviceLock              sync.Mutex
	Locker                  *LightLock
	Logger                  logger.Logger
	available               bool
	lastColor               string
	on                      bool
	lock                    sync.RWMutex
	LastProgrammedColorName string
	LastProgrammedColor     string
	activeScene             string
}

var Lights = make(map[string]*Light)
var LightGroups = make(map[string][]*Light)

func (l *Light) SetLastProgrammedColor(name, color string) {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.LastProgrammedColorName = name
	l.LastProgrammedColor = color
}

func (l *Light) GetLastProgrammedColor() (string, string) {
	l.lock.RLock()
	defer l.lock.RUnlock()
	return l.LastProgrammedColorName, l.LastProgrammedColor
}

func (l *Light) ActiveScene() string {
	l.lock.RLock()
	defer l.lock.RUnlock()
	return l.activeScene
}

func (l *Light) SetActiveScene(scene string) {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.activeScene = scene
}

func (l *Light) Available() bool {
	l.lock.RLock()
	defer l.lock.RUnlock()
	return l.available
}

func (l *Light) SetAvailable(available bool) {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.available = available
}
func (l *Light) LastColor() string {
	l.lock.RLock()
	defer l.lock.RUnlock()
	return l.lastColor
}

func (l *Light) SetLastColor(color string) {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.lastColor = color
}

func (l *Light) On() bool {
	l.lock.RLock()
	defer l.lock.RUnlock()
	return l.on
}

func (l *Light) SetOn(on bool) {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.on = on
}

func AddLight(name, addr string) error {

	addr = strings.ToLower(strings.TrimSpace(addr))
	addressIsIp := true

	var ip, mac string

	if net.ParseIP(addr) != nil {
		addressIsIp = true
		ip = addr
	} else {
		if macReg.MatchString(addr) {
			addressIsIp = false
			mac = addr
		} else {
			return fmt.Errorf("invalid address %q", addr)
		}
	}

	for existingName, existingLight := range Lights {
		if addressIsIp && existingLight.device != nil && existingLight.device.IP() == addr {
			return fmt.Errorf("light with Address %q already exists as %q", addr, existingName)
		}
		if !addressIsIp && existingLight.mac == addr {
			return fmt.Errorf("light with Address %q already exists as %q", addr, existingName)
		}
		if existingName == name {
			return errors.New("duplicate light name")
		}
	}

	var device *wiz.Device
	var err error
	if addressIsIp {
		device, err = wiz.NewDevice(addr)
		if err != nil {
			return err
		}
		device.SetName(name).WithLogger(common.Logger.GetLogger(fmt.Sprintf("device-%s", name)))
	}

	log := common.Logger.GetLogger(fmt.Sprintf("device-%s", name))

	Lights[name] = &Light{
		device:    device,
		ip:        ip,
		mac:       mac,
		lastColor: "",
		available: false,
		Logger:    log,
		Locker:    newLocker(name),
		name:      name,
	}

	return nil
}

func AddLightGroup(name string, members []string) error {
	if _, exists := LightGroups[name]; exists {
		return fmt.Errorf("light group %q already exists", name)

	}

	if _, exists := Lights[name]; exists {
		return fmt.Errorf("light group %q already exists as a light, all lights and light groups must have distinct names", name)
	}

	devices := make([]*Light, 0, len(members))
	for _, member := range members {
		device, exists := Lights[member]
		if !exists {
			return fmt.Errorf("light %q does not exist", member)
		}
		devices = append(devices, device)
	}

	LightGroups[name] = devices
	return nil
}

func IsValidLightOrGroup(name string) bool {
	_, isLight := Lights[name]
	_, isGroup := LightGroups[name]
	return isLight || isGroup
}

func GetLightOrGroup(name string) ([]*Light, error) {
	if light, isLight := Lights[name]; isLight {
		return []*Light{light}, nil
	}
	if group, isGroup := LightGroups[name]; isGroup {
		return group, nil
	}
	return nil, fmt.Errorf("no light or group found with name %q", name)
}

func (l *Light) GetName() string {
	return l.name
}

// GetDevice returns the underlying wiz device
// caution: may be nil if discovery is used
func (l *Light) GetDevice() *wiz.Device {
	l.deviceLock.Lock()
	defer l.deviceLock.Unlock()
	return l.device
}

// SetDevice is used by discovery to associated a discovered device with this light
func (l *Light) SetDevice(dev *wiz.Device) {
	l.deviceLock.Lock()
	defer l.deviceLock.Unlock()
	l.device = dev.SetName(l.name).WithLogger(l.Logger)
	l.ip = dev.IP()
}

// RequiresDiscovery returns true if the device was configured with a MAC address
// Such devices require discovery assistance to obtain the IP so we can talk to them
func (l *Light) RequiresDiscovery() bool {
	return l.mac != ""
}

// GetAddress returns the IP and MAC address of the light
// At least one will be defined but either may be empty
// If mac is defined, device will participate in discovery
func (l *Light) GetAddress() (ip, mac string) {
	// N.B. because SetDevice might be updating the IP, we will use deviceLock to
	// ensure against concurrent updates to the IP by SetDevice
	l.deviceLock.Lock()
	defer l.deviceLock.Unlock()
	return l.ip, l.mac
}

// GetAllLightsDeviceGroup returns a wiz device group with devices
// for all lights. This will exclude any lights
// that were configured with MAC address and for which we have not yet discovered an IP
func GetAllLightsDeviceGroup() *wiz.DeviceGroup {
	group := wiz.NewDeviceGroup()
	for _, light := range Lights {
		if light.device != nil {
			group.AddDevice(light.device)
		}
	}
	return group
}
