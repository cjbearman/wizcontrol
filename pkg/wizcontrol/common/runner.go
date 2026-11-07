package common

import (
	"fmt"

	"github.com/cjbearman/logger"
)

// Runner provides a consistent interface for running our various background service tasks
type Runner interface {
	// Start service
	Start()

	// Stop service
	Stop()

	// Stopped returns a channel that will be closed when the runner has stopped
	Stopped() <-chan struct{}
}

// BaseRunner is a concrete implementation of the Runner interface that provides basic lifecycle management using channels for stop and close signals.
type BaseRunner struct {
	StopSignalled chan struct{}
	Closed        chan struct{}
	Exec          func()
	logger        logger.Logger
}

// NewRunner creates a new BaseRunner instance with the specified name.
// The Exec method must be set before calling Start.
func NewRunner(name string) *BaseRunner {
	return &BaseRunner{
		StopSignalled: make(chan struct{}),
		Closed:        make(chan struct{}),
		logger:        Logger.GetLogger(fmt.Sprintf("runner-%s", name)),
	}
}

// Start begins execution of the runner by invoking the Exec function in a separate goroutine.
func (r *BaseRunner) Start() {
	r.logger.Info("Starting runner")
	go r.Exec()
}

// Stop signals the runner to stop by closing the StopSignalled channel.
func (r *BaseRunner) Stop() {
	r.logger.Info("Stopping runner by closing StopSignalled channel")
	close(r.StopSignalled)
}

// Stopped returns a channel that will be closed when the runner has stopped
func (r *BaseRunner) Stopped() <-chan struct{} {
	return r.Closed
}
