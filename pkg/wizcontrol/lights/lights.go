package lights

import (
	"errors"
	"fmt"

	"github.com/cjbearman/framework/pkg/logger"
	"github.com/cjbearman/wiz/pkg/wiz"
	"github.com/cjbearman/wizcontrol2/pkg/wizcontrol/common"
)

type Light struct {
	Light     *wiz.Device
	Logger    logger.Logger
	Available bool
	LastColor string
	On        bool
}

var Lights = make(map[string]*Light)
var LightGroups = make(map[string][]*Light)

func AddLight(name, ip string) error {
	for existingName, existingLight := range Lights {
		if existingLight.Light.IP() == ip {
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
		Light:     device.SetName(name).WithLogger(log),
		LastColor: "",
		Available: false,
		Logger:    log,
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
