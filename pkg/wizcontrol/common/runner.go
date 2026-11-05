package common

import (
	"fmt"

	"github.com/cjbearman/logger"
)

type Runner interface {
	Start()
	Stop()
	Stopped() <-chan struct{}
}

// Runner is a parent object for lifecycle management using channels for stop and close signals
type BaseRunner struct {
	StopSignalled chan struct{}
	Closed        chan struct{}
	Exec          func()
	logger        logger.Logger
}

// Creates a new runner
// Returns a pointer to the newly created Runner instance
// Exec method must be set before calling start
func NewRunner(name string) *BaseRunner {
	return &BaseRunner{
		StopSignalled: make(chan struct{}),
		Closed:        make(chan struct{}),
		logger:        Logger.GetLogger(fmt.Sprintf("runner-%s", name)),
	}
}

// Start runs the exec method in a goroutine
func (r *BaseRunner) Start() {
	r.logger.Info("Starting runner")
	go r.Exec()
}

// Stop signals exit
// Deriving class can override this if needed
func (r *BaseRunner) Stop() {
	r.logger.Info("Stopping runner by closing StopSignalled channel")
	close(r.StopSignalled)
}

// Stopped returns a channel that will be closed when the runner has stopped
func (r *BaseRunner) Stopped() <-chan struct{} {
	return r.Closed
}
