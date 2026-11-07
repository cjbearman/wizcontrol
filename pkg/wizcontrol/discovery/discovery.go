// Package discovery handles discovery of lights by their mac addresses
// in case any light was configured with mac and not IP
package discovery

import (
	"strings"
	"time"

	"github.com/cjbearman/wiz/pkg/wiz"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/common"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/config"
)

// Instance is the singleton instance of the discovery service
var Instance *Discovery

// Discovery represents the discovery service responsible for finding lights by their MAC addresses
type Discovery struct {
	*common.BaseRunner
	discoveryBroadcastAddr string
	discoveryInterval      time.Duration
	discoveryPeriod        time.Duration
	cnf                    *config.Config
}

// NewDiscovery creates a new instance of the discovery service
func NewDiscovery(cnf *config.Config) *Discovery {
	d := &Discovery{
		BaseRunner:             common.NewRunner("discovery"),
		discoveryBroadcastAddr: cnf.Discovery.BroadcastAddress(),
		discoveryInterval:      cnf.Discovery.Interval(),
		discoveryPeriod:        cnf.Discovery.Period(),
		cnf:                    cnf,
	}

	d.Exec = d.run
	return d
}

// run is the main loop for the discovery service, periodically performing discovery cycles
func (d *Discovery) run() {
	logger := common.Logger.GetLogger("discovery").WithField("interval", d.discoveryInterval).
		WithField("period", d.discoveryPeriod).WithField("broadcast", d.discoveryBroadcastAddr)

	// Ensure the Closed channel is closed when the discovery service exits
	defer close(d.Closed)

	logger.Info("Discovery started")

	// A ticker is used to handle the periodic discovery cycles
	ticker := time.NewTicker(d.discoveryInterval)
	defer ticker.Stop()

	for {
		// Set up a wiz discovery component
		disc := wiz.Discover().BroadcastInterval(d.discoveryInterval).EndAfter(d.discoveryPeriod).WithCallback(d.callback)

		// A broadcast address can be specified to direct discovery traffic to a specific network interface
		if d.discoveryBroadcastAddr != "" {
			disc.BroadcastAddress(d.discoveryBroadcastAddr)
		}

		// Perform the discovery cycle
		logger.Debugf("Starting a discovery cycle")
		devs, err := disc.GetDiscoveredDevices()
		if err != nil {
			logger.Errorf("Discovery failed: %v", err)
		} else {
			logger.Infof("Discovered %d total devices", len(devs))
		}

		// Reset the ticker to ensure the next discovery happens after the configured interval
		ticker.Reset(d.discoveryInterval)

		// Wait for the next tick or stop signal before starting the next discovery cycle
		select {
		case <-ticker.C:
		case <-d.StopSignalled:
			// Time to shut down
			logger.Infof("Stop signalled, terminating discovery")
			return
		}
	}
}

// callback is called by the wiz discovery service for every device discovered
func (d *Discovery) callback(dev *wiz.Device) {
	logger := common.Logger.GetLogger("discovery")
	logger.Debugf("Device discovery inbound for MAC %s IP %s", dev.MAC(), dev.IP())

	// Go through each configured light to see if we have one configured for discovery at this mac
	for _, l := range d.cnf.Lights {
		mac := l.DiscoveryMAC()
		if mac != "" {
			// Device has a MAC and is enabled for discovery
			if strings.EqualFold(mac, dev.MAC()) {
				// Matches this device

				updateDevice := false
				existingDevice := l.Device()

				if existingDevice == nil {
					// The light has no wiz device associated yet, so this is a first discovery of the device
					logger.Infof("First discovery of configured device with MAC %s at IP %s", dev.MAC(), dev.IP())
					updateDevice = true
				} else {
					// This light has previously been discovered, but it's IP may have changed
					if existingDevice.IP() != dev.IP() {
						// The IP has changed, so it will need an updated wiz device
						logger.Infof("Device with MAC %s has new IP %s (was %s)", dev.MAC(), dev.IP(), existingDevice.IP())
						updateDevice = true
					} else {
						// IP unchanged, so existing wiz device is good
						logger.Debugf("Rediscovered device with MAC %s at unchanged IP %s", dev.MAC(), dev.IP())
					}
				}
				// If updateDevice is true, associate the discovered device with the light
				if updateDevice {
					l.SetDevice(dev)
				}
			}
		}
	}
}
