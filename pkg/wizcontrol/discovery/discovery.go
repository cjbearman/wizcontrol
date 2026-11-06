package discovery

import (
	"strings"
	"time"

	"github.com/cjbearman/wiz/pkg/wiz"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/common"
	dc "github.com/cjbearman/wizcontrol/pkg/wizcontrol/component/discovery"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/lights"
)

var Instance *Discovery

type Discovery struct {
	*common.BaseRunner
	discoveryBroadcastAddr string
	discoveryInterval      time.Duration
	discoveryPeriod        time.Duration
}

func NewDiscovery(cnf *dc.Discovery) *Discovery {
	d := &Discovery{
		BaseRunner:             common.NewRunner("discovery"),
		discoveryBroadcastAddr: cnf.BroadcastAddress(),
		discoveryInterval:      cnf.Interval(),
		discoveryPeriod:        cnf.Period(),
	}

	d.Exec = d.run
	return d
}

func (d *Discovery) run() {
	logger := common.Logger.GetLogger("discovery").WithField("interval", d.discoveryInterval).
		WithField("period", d.discoveryPeriod).WithField("broadcast", d.discoveryBroadcastAddr)
	defer close(d.Closed)
	logger.Info("discovery started")

	ticker := time.NewTicker(d.discoveryInterval)
	defer ticker.Stop()

	for {
		disc := wiz.Discover().BroadcastInterval(d.discoveryInterval).EndAfter(d.discoveryPeriod).WithCallback(d.callback)
		if d.discoveryBroadcastAddr != "" {
			disc.BroadcastAddress(d.discoveryBroadcastAddr)
		}

		// Do a discovery cycle, d.callback will receive information of all discovered devices
		logger.Debugf("starting a discovery cycle")
		devs, err := disc.GetDiscoveredDevices()
		if err != nil {
			logger.Errorf("Discovery failed: %v", err)
		} else {
			logger.Infof("Discovered %d total devices", len(devs))
		}

		// Reset the ticker to ensure the next discovery happens after the configured interval
		ticker.Reset(d.discoveryInterval)

		select {
		case <-ticker.C:
		case <-d.StopSignalled:
			logger.Infof("Stop signalled, terminating discovery")
			return
		}
	}
}

func (d *Discovery) callback(dev *wiz.Device) {
	logger := common.Logger.GetLogger("discovery")
	logger.Debugf("Device discovery inbound for MAC %s IP %s", dev.MAC(), dev.IP())

	for _, l := range lights.GetLights() {
		_, mac := l.GetAddress()
		if mac != "" {
			// Device has a MAC and is enabled for discovery

			if strings.EqualFold(mac, dev.MAC()) {
				updateDevice := false
				// There is a device that is set to use this MAC address
				existingDevice := l.GetDevice()

				if existingDevice == nil {
					logger.Infof("First discovery of configured device with MAC %s at IP %s", dev.MAC(), dev.IP())
					updateDevice = true
				} else {
					if existingDevice.IP() != dev.IP() {
						logger.Infof("Device with MAC %s has new IP %s (was %s)", dev.MAC(), dev.IP(), existingDevice.IP())
						updateDevice = true
					} else {
						logger.Debugf("Rediscovered device with MAC %s at unchanged IP %s", dev.MAC(), dev.IP())
					}
				}
				if updateDevice {
					l.SetDevice(dev)
				}
			}
		}
	}

}
