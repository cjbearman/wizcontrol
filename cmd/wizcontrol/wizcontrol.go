package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"syscall"
	"time"

	"github.com/cjbearman/logger"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/common"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/config"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/daemon"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/discovery"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/mqtt"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/server"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/state"
)

// main is the main runner for wizcontrol
func main() {
	// Command line processing
	logLevel := flag.String("log", "info", "Set the logging level")
	cfg := flag.String("config", "config.yaml", "Path to the configuration file")
	wc := flag.Bool("wc", false, "List available pre-defined wiz colors")
	flag.Parse()

	// Set up logging based on the command line flag for log level
	level, err := logger.LevelFromString(*logLevel)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid log level %q\n", *logLevel)
		os.Exit(1)
	}
	common.Logger = logger.NewLoggerFactory(level, logger.NewColorStringFormatter(), logger.NewSyncWriter(os.Stdout)).GetLogger("root")

	DaemonStartTime := time.Now().In(common.Location)
	DaemonRestartTime := DaemonStartTime

	for {
		// Load the configuration file specified by the command line flag
		parsed, err := config.Load(*cfg)
		if err != nil {
			common.Logger.Errorf("Failed to load config: %v", err)
			os.Exit(1)
		}

		// Build the runtime configuration
		finalConfig, err := config.Build(parsed)
		if err != nil {
			common.Logger.Errorf("Failed to build config: %v", err)
			os.Exit(1)
		}

		// Set our system wide timezone and add timestamping to the logger in the same timezone
		common.Location = finalConfig.System.Timezone()
		common.Logger = common.Logger.WithField("time", logger.NewTimestamper(time.DateTime+".000000", common.Location))
		common.Logger.Infof("Operating in location: %v", common.Location)

		// If the user requested to list available pre-defined wiz colors, do that and exit
		if *wc {
			for name := range finalConfig.WizColors {
				fmt.Printf("%s\n", name)
			}
			os.Exit(0)
		}

		// Run the application as a daemon with the final configuration
		finalConfig.StartTime = DaemonStartTime
		finalConfig.RestartTime = DaemonRestartTime
		runAsDaemon(finalConfig)
		DaemonRestartTime = time.Now().In(common.Location)
	}
}

// runAsDaemon runs the wizcontrol service through completion
func runAsDaemon(cnf *config.Config) {
	logger := common.Logger.GetLogger("runner")
	logger.Info("Runner starting")

	// We need to catch CTRL-C and other termination signals
	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, os.Interrupt, syscall.SIGTERM)

	// Load any state recovery file
	if err := state.RecoverState(cnf); err != nil {
		common.Logger.Warnf("Failed to recover state: %v", err)
	}

	// Start the main daemon
	logger.Info("Starting daemon")
	daemon.Instance = daemon.NewDaemon(cnf)
	daemon.Instance.Start()

	// Start the web server, if configured
	if cnf.HTTP != nil {
		logger.Info("Starting HTTP")
		server.Instance = server.NewServer(cnf)
		server.Instance.Start()
	} else {
		logger.Info("HTTP not configured")
	}

	// And MQTT, if configured
	if cnf.MQTT != nil {
		logger.Info("Starting MQTT")
		mqtt.Instance = mqtt.NewMqtt(cnf)
		mqtt.Instance.Start()
	} else {
		logger.Info("MQTT not configured")
	}

	// If one or more devices were configured on a Mac basis, then we need to start discovery to find them
	if cnf.Lights.DiscoveryRequired() {
		logger.Info("One or more lights require discovery, discovery service will be started")
		discovery.Instance = discovery.NewDiscovery(cnf)
		discovery.Instance.Start()
	} else {
		logger.Info("No devices are configured with MAC address, so discovery is not needed and discovery service will not run")
	}

	// We may or may not have running services, but we need a valid channel
	// for all to listen to (even if there is no service backing it)
	mqttStopCh := guaranteeChannel(mqtt.Instance)
	discoveryStopCh := guaranteeChannel(discovery.Instance)
	daemonStopCh := guaranteeChannel(daemon.Instance)
	serverStopCh := guaranteeChannel(server.Instance)

	restartRequested := false

	// Everything is now running, wait for something to stop or for a termination signal
	// and shutdown things as needed
	select {
	case <-mqttStopCh:
		logger.Warn("MQTT has stopped, stopping daemon and server")
		stop(daemon.Instance)
		stop(server.Instance)
		stop(discovery.Instance)

	case <-daemonStopCh:
		logger.Warn("Daemon has stopped, stopping other tasks")
		stop(server.Instance)
		stop(mqtt.Instance)
		stop(discovery.Instance)

	case <-serverStopCh:
		logger.Warn("Web server has stopped, stopping other tasks")
		stop(mqtt.Instance)
		stop(daemon.Instance)
		stop(discovery.Instance)

	case <-discoveryStopCh:
		logger.Warn("Discovery service has stopped, stopping other tasks")
		stop(daemon.Instance)
		stop(server.Instance)
		stop(mqtt.Instance)

	case sig := <-sigch:
		logger.Warnf("Caught signal %v, stopping tasksr", sig)
		stop(daemon.Instance)
		stop(server.Instance)
		stop(mqtt.Instance)
		stop(discovery.Instance)

	case <-cnf.RestartCh:
		logger.Warn("Restart requested, stopping other tasks")
		stop(daemon.Instance)
		stop(server.Instance)
		stop(mqtt.Instance)
		stop(discovery.Instance)
		restartRequested = true
	}

	// Everything should now be stopping, wait for all services to confirm they have stopped
	// order isn't important
	waitStopped(daemon.Instance)
	waitStopped(server.Instance)
	waitStopped(mqtt.Instance)
	waitStopped(discovery.Instance)

	// Close out our message bus, ensuring it gets drained for cleanup
	cnf.Bus.Close()
	logger.Info("Runner terminated")

	// That's all folks
	if restartRequested {
		logger.Warnf("Restarting...")
		return
	}

	// Was not a restart, so we're done
	os.Exit(0)

}

// stop will stop any service that is not nil
func stop(service common.Runner) {
	if !isNil(service) {
		service.Stop()
	}
}

// waitStopped will wait for a service to stop or return immediately if nil
func waitStopped(service common.Runner) {
	if !isNil(service) {
		<-service.Stopped()
	}
}

// guaranteeChannel will return the stopped channel for a service or a dummy channel if nil
func guaranteeChannel(service common.Runner) <-chan struct{} {
	if !isNil(service) {
		return service.Stopped()
	}
	ch := make(chan struct{})
	return ch
}

// isNil returns true if the given interface is nil or points to a nil value
// because checking an interface directly against nil is problematic
func isNil(i any) bool {
	if i == nil {
		return true
	}

	v := reflect.ValueOf(i)
	switch v.Kind() {
	// Only call IsNil on types that can actually be nil
	case reflect.Pointer, reflect.Map, reflect.Slice, reflect.Chan, reflect.Func, reflect.Interface:
		return v.IsNil()
	}

	return false
}
