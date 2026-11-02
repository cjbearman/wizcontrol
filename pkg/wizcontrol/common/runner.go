package common

// Runner is a parent object for lifecycle management using channels for stop and close signals
type Runner struct {
	StopSignalled chan struct{}
	Closed        chan struct{}
	Exec          func()
}

// Creates a new runner
// Returns a pointer to the newly created Runner instance
// Exec method must be set before calling start
func NewRunner() *Runner {
	return &Runner{
		StopSignalled: make(chan struct{}),
		Closed:        make(chan struct{}),
	}
}

// Start runs the exec method in a goroutine
func (r *Runner) Start() {
	go r.Exec()
}

// Stop signals exit
// Deriving class can override this if needed
func (r *Runner) Stop() {
	close(r.StopSignalled)
}

// Stopped returns a channel that will be closed when the runner has stopped
func (r *Runner) Stopped() <-chan struct{} {
	return r.Closed
}
