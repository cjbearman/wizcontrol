package discovery

import (
	"fmt"
	"time"
)

type Discovery struct {
	broadcastAddress string
	interval         time.Duration
	period           time.Duration
}

const (
	DefaultInterval = 30 * time.Second
	DefaultPeriod   = 5 * time.Second
)

func Build(cfg ConfigYAML) Discovery {
	return Discovery{
		broadcastAddress: cfg.BroadcastAddress,
		interval: func() time.Duration {
			d, _ := time.ParseDuration(cfg.Interval)
			if d == 0 {
				return DefaultInterval
			}
			return d
		}(),
		period: func() time.Duration {
			d, _ := time.ParseDuration(cfg.Period)
			if d == 0 {
				return DefaultPeriod
			}
			return d
		}(),
	}
}

func (d Discovery) String() string {
	return fmt.Sprintf("Discovery: broadcast_address=%s, interval=%s, period=%s", d.broadcastAddress, d.interval, d.period)
}

func (d Discovery) BroadcastAddress() string {
	return d.broadcastAddress
}

func (d Discovery) Interval() time.Duration {
	return d.interval
}

func (d Discovery) Period() time.Duration {
	return d.period
}
