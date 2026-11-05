package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"slices"
	"syscall"
	"time"

	"github.com/cjbearman/logger"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/calendar"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/colors"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/common"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/config"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/daemon"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/discovery"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/lights"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/mqtt"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/operations"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/scenes"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/server"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/state"
)

func main() {
	logLevel := flag.String("log", "info", "Set the logging level")
	cfg := flag.String("config", "config.yaml", "Path to the configuration file")
	apply := flag.String("apply", "", "Apply scene")
	f := flag.Bool("f", false, "Force apply scene")
	d := flag.Bool("d", false, "Run as daemon")
	cd := flag.Bool("cd", false, "Enable calendar debug")
	presets := flag.Bool("presets", false, "List available preset colors")

	flag.Parse()

	if *presets {
		// List presets before anything else, because if we load config, we'll get other non-preset colors defined
		presetsList := []string{}
		for preset := range colors.GetColors() {
			presetsList = append(presetsList, preset)
		}

		slices.Sort(presetsList)

		for _, preset := range presetsList {
			fmt.Println(preset)
		}
		os.Exit(0)
	}

	level, err := logger.LevelFromString(*logLevel)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid log level %q\n", *logLevel)
		os.Exit(1)
	}

	common.Logger = logger.NewLoggerFactory(level, logger.NewColorStringFormatter(), logger.NewSyncWriter(os.Stdout)).GetLogger("root")

	if _, err := config.Load(*cfg); err != nil {
		common.Logger.Errorf("Failed to load config: %v", err)
		os.Exit(1)
	}
	common.Logger = common.Logger.WithField("time", logger.NewTimestamper(time.DateTime+".000000", common.Location))

	common.Logger.Infof("Operating in location: %v", common.Location)

	if *apply != "" {

		_, found := scenes.GetScenes()[*apply]
		if !found {
			common.Logger.Errorf("Scene %q not found", *apply)
			os.Exit(1)
		}
		operations.UpdateAvailability()
		operations.ApplyScenes([]string{*apply}, *f, false, false)
		os.Exit(0)
	}

	if *d {
		runAsDaemon()
	}
	if *cd {
		calendar.CalendarDebug = true
	}
}

func runAsDaemon() {
	logger := common.Logger.GetLogger("runner")
	logger.Info("Runner starting")
	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, os.Interrupt, syscall.SIGTERM)

	// Load any state recovery file
	if err := state.RecoverState(); err != nil {
		common.Logger.Warnf("Failed to recover state: %v", err)
	}

	// Start the main daemon and web server
	daemon.Instance = daemon.NewDaemon()
	daemon.Instance.Start()
	server.Instance = server.NewServer()
	server.Instance.Start()

	// And MQTT if configured
	mqtt.Instance = mqtt.NewMqtt()
	if mqtt.MqttURL != "" {
		mqtt.Instance.Start()
	}

	// If one or more devices were configured on a Mac basis, then we need to run discovery to find them
	for _, light := range lights.GetLights() {
		if light.RequiresDiscovery() {
			logger.Info("One or more lights require discovery, discovery service will be started")
			discovery.Instance = discovery.NewDiscovery()
			discovery.Instance.Start()
			break
		}
	}

	if discovery.Instance == nil {
		logger.Info("No devices are configured with MAC address, so discovery is not needed and discovery service will not run")
	}

	// We may or may not have running services, but we need a valid channel
	// for all to listen to (even if there is no service backing it)
	mqttStopCh := guaranteeChannel(mqtt.Instance)
	discoveryStopCh := guaranteeChannel(discovery.Instance)
	daemonStopCh := guaranteeChannel(daemon.Instance)
	serverStopCh := guaranteeChannel(server.Instance)

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
		stop(server.Instance)
		stop(mqtt.Instance)
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
	}

	waitStopped(daemon.Instance)
	waitStopped(server.Instance)
	waitStopped(mqtt.Instance)
	waitStopped(discovery.Instance)
	logger.Info("Runner terminated")
	os.Exit(0)

}

func stop(service common.Runner) {
	if !isNil(service) {
		service.Stop()
	}
}

func waitStopped(service common.Runner) {
	if !isNil(service) {
		<-service.Stopped()
	}
}

func guaranteeChannel(service common.Runner) <-chan struct{} {
	if !isNil(service) {
		return service.Stopped()
	}
	ch := make(chan struct{})
	return ch
}

func isNil(i any) bool {
	if i == nil {
		return true
	}

	v := reflect.ValueOf(i)
	switch v.Kind() {
	// Only call IsNil on types that can actually be nil
	case reflect.Ptr, reflect.Map, reflect.Slice, reflect.Chan, reflect.Func, reflect.Interface:
		return v.IsNil()
	}

	return false
}
