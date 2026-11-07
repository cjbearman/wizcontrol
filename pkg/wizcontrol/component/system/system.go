package system

import (
	"fmt"
	"time"
)

// System represents the system configuration, including timezone, evaluation period, recheck period, and Wiz color prefix.
type System struct {
	timezone         *time.Location
	evaluationPeriod time.Duration
	recheckPeriod    time.Duration
	wizColorPrefix   string
}

const (
	// DefaultEvaluationPeriod is the default duration for the evaluation period if not specified.
	DefaultEvaluationPeriod = 15 * time.Second

	// DefaultRecheckPeriod is the default duration for the recheck period if not specified.
	DefaultRecheckPeriod = 5 * time.Minute
)

// Build constructs a System instance from the provided ConfigYAML, applying default values where necessary.
func Build(cfg ConfigYAML) System {
	return System{
		timezone: func() *time.Location {
			if cfg.Timezone == "" {
				return time.Local
			}
			loc, _ := time.LoadLocation(cfg.Timezone)
			return loc
		}(),
		evaluationPeriod: func() time.Duration {
			d, err := time.ParseDuration(cfg.EvaluationPeriod)
			if err != nil || d == 0 {
				return DefaultEvaluationPeriod
			}
			return d
		}(),
		recheckPeriod: func() time.Duration {
			d, err := time.ParseDuration(cfg.RecheckPeriod)
			if err != nil || d == 0 {
				return DefaultRecheckPeriod
			}
			return d
		}(),
		wizColorPrefix: cfg.WizColorPrefix,
	}
}

// String returns a string representation of the System instance.
func (s System) String() string {
	return fmt.Sprintf("System: timezone=%s, evaluation_period=%s, recheck_period=%s, wiz_color_prefix=%s",
		s.timezone, s.evaluationPeriod, s.recheckPeriod, s.wizColorPrefix)
}

// Timezone returns the timezone.
func (s System) Timezone() *time.Location {
	return s.timezone
}

// EvaluationPeriod returns the evaluation period.
func (s System) EvaluationPeriod() time.Duration {
	return s.evaluationPeriod
}

// RecheckPeriod returns the recheck period.
func (s System) RecheckPeriod() time.Duration {
	return s.recheckPeriod
}

// WizColorPrefix returns the Wiz color prefix.
func (s System) WizColorPrefix() string {
	return s.wizColorPrefix
}
