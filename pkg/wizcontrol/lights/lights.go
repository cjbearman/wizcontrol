package lights

import (
	"errors"
	"fmt"

	"github.com/cjbearman/wiz/pkg/wiz"
	"github.com/cjbearman/wizcontrol2/pkg/wizcontrol/common"
)

var Lights = make(map[string]*wiz.Device)
var LightGroups = make(map[string][]*wiz.Device)

func AddLight(name, ip string) error {
	for existingName, existingDevice := range Lights {
		if existingDevice.IP() == ip {
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

	Lights[name] = device.SetName(name).WithLogger(common.Logger.GetLogger(fmt.Sprintf("device-%s", name)))
	return nil
}

func AddLightGroup(name string, members []string) error {
	if _, exists := LightGroups[name]; exists {
		return fmt.Errorf("light group %q already exists", name)

	}

	if _, exists := Lights[name]; exists {
		return fmt.Errorf("light group %q already exists as a light, all lights and light groups must have distinct names", name)
	}

	devices := make([]*wiz.Device, 0, len(members))
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
