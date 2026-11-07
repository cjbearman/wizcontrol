package light

import "github.com/cjbearman/wiz/pkg/wiz"

// Lights represents a collection of lights keyed by their name.
type Lights map[string]*Light

// getLightByIP returns the light with the specified IP address, if any.
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

// getLightsByDiscoveryMac returns the light with the specified discovery MAC address, if any.
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

// DiscoveryRequired returns true if any light in the collection requires discovery mode.
// which is determined by the light having a discovyery MAC address configured.
func (l Lights) DiscoveryRequired() bool {
	for _, light := range l {
		if light.UseDiscovery() {
			return true
		}
	}
	return false
}

// GetAllLightsDeviceGroup returns a wiz device group containing all wiz devices
// for all lights that are:
// 1. Configured by IP, or
// 2. Configured by MAC and have been discovered
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
