package lights

import (
	"errors"
	"fmt"
	"sync"

	"github.com/cjbearman/framework/pkg/logger"
	"github.com/cjbearman/wiz/pkg/wiz"
	"github.com/cjbearman/wizcontrol2/pkg/wizcontrol/common"
)

type Light struct {
	Device                  *wiz.Device
	Locker                  *LightLock
	Logger                  logger.Logger
	available               bool
	lastColor               string
	on                      bool
	lock                    sync.RWMutex
	LastProgrammedColorName string
	LastProgrammedColor     string
}

var Lights = make(map[string]*Light)
var LightGroups = make(map[string][]*Light)
var AllDevicesGroup *wiz.DeviceGroup = wiz.NewDeviceGroup()

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

func AddLight(name, ip string) error {
	for existingName, existingLight := range Lights {
		if existingLight.Device.IP() == ip {
			return fmt.Errorf("light with IP %q already exists as %q", ip, existingName)
		}
		if existingName == name {
			return errors.New("duplicate light name")
		}
	}

	device, err := wiz.NewDevice(ip)
	if err != nil {
		return err
	}

	device.SetName(name).WithLogger(common.Logger.GetLogger(fmt.Sprintf("device-%s", name)))

	log := common.Logger.GetLogger(fmt.Sprintf("device-%s", name))

	Lights[name] = &Light{
		Device:    device.SetName(name).WithLogger(log),
		lastColor: "",
		available: false,
		Logger:    log,
		Locker:    newLocker(name),
	}
	AllDevicesGroup.AddDevice(device)
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
