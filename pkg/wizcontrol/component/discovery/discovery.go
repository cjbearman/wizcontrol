package discovery

import (
	"fmt"
	"time"
)

// Discovery represents configuration for discovery operations
type Discovery struct {
	broadcastAddress string
	interval         time.Duration
	period           time.Duration
}

const (
	// DefaultInterval is the default interval at which discovery runs are initiated
	DefaultInterval = 30 * time.Second

	// DefaultPeriod is the default interval to wait for responses after sending discovery packets
	DefaultPeriod = 5 * time.Second
)

// Build constructs a Discovery instance from the given ConfigYAML
func Build(cfg ConfigYAML) Discovery {
	return Discovery{
		broadcastAddress: cfg.BroadcastAddress,
		interval: func() time.Duration {
			if cfg.Interval == "" {
				return DefaultInterval
			}
			d, _ := time.ParseDuration(cfg.Interval) // N.B. already validated when set
			return d
		}(),
		period: func() time.Duration {
			if cfg.Period == "" {
				return DefaultPeriod
			}
			d, _ := time.ParseDuration(cfg.Period) // N.B. already validated when set
			return d
		}(),
	}
}

// String returns a string representation of the Discovery instance
func (d Discovery) String() string {
	return fmt.Sprintf("Discovery: broadcast_address=%s, interval=%s, period=%s", d.broadcastAddress, d.interval, d.period)
}

// BroadcastAddress returns the broadcast address of the Discovery instance
func (d Discovery) BroadcastAddress() string {
	return d.broadcastAddress
}

// Interval returns the interval at which discovery runs are initiated
func (d Discovery) Interval() time.Duration {
	return d.interval
}

// Period returns the period to wait for responses after sending discovery packets
func (d Discovery) Period() time.Duration {
	return d.period
}
