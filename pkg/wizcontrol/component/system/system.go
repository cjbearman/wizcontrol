package system

import (
	"fmt"
	"time"
)

type System struct {
	timezone         *time.Location
	evaluationPeriod time.Duration
	recheckPeriod    time.Duration
}

const (
	DefaultEvaluationPeriod = 15 * time.Second
	DefaultRecheckPeriod    = 5 * time.Minute
)

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
	}
}

func (s System) String() string {
	return fmt.Sprintf("System: timezone=%s, evaluation_period=%s, recheck_period=%s",
		s.timezone, s.evaluationPeriod, s.recheckPeriod)
}

func (s System) Timezone() *time.Location {
	return s.timezone
}

func (s System) EvaluationPeriod() time.Duration {
	return s.evaluationPeriod
}

func (s System) RecheckPeriod() time.Duration {
	return s.recheckPeriod
}
