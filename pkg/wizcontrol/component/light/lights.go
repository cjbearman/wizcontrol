package light

import "github.com/cjbearman/wiz/pkg/wiz"

type Lights map[string]*Light

func (l Lights) getLightByIP(ip string) *Light {
	if ip == "" {
		return nil
	}
	for _, light := range l {
		if light.ip == ip {
			return light
		}
	}
	return nil
}

func (l Lights) getLightsByDiscoveryMac(mac string) *Light {
	for _, light := range l {
		if light.discoveryMAC == "" {
			continue
		}
		if light.discoveryMAC == mac {
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

func (l Lights) DiscoveryRequired() bool {
	for _, light := range l {
		if light.UseDiscovery() {
			return true
		}
	}
	return false
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
