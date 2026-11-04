package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"slices"
	"syscall"
	"time"

	"github.com/cjbearman/framework/pkg/logger"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/calendar"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/colors"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/common"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/config"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/daemon"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/mqtt"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/operations"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/scenes"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/server"
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
		for preset := range colors.Colors {
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

		_, found := scenes.Scenes[*apply]
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

	// Start the main daemon and web server
	daemon.Daemon = daemon.NewDaemon()
	daemon.Daemon.Start()
	server.Server = server.NewServer()
	server.Server.Start()

	// And MQTT if configured
	mqtt.MqttInstance = mqtt.NewMqtt()
	if mqtt.MqttURL != "" {
		mqtt.MqttInstance.Start()
	}

	select {

	case <-mqtt.MqttInstance.Stopped():
		logger.Warn("MQTT has stopped, stopping daemon and server")
		daemon.Daemon.Stop()
		server.Server.Stop()

	case <-daemon.Daemon.Stopped():
		logger.Warn("Daemon has stopped, stopping other tasks")
		server.Server.Stop()
		if mqtt.MqttURL != "" {
			mqtt.MqttInstance.Stop()
		}

	case <-server.Server.Stopped():
		logger.Warn("Web server has stopped, stopping other tasks")
		daemon.Daemon.Stop()
		if mqtt.MqttURL != "" {
			mqtt.MqttInstance.Stop()
		}
	case sig := <-sigch:
		logger.Warnf("Caught signal %v, stopping tasksr", sig)
		daemon.Daemon.Stop()
		server.Server.Stop()
		if mqtt.MqttURL != "" {
			mqtt.MqttInstance.Stop()
		}
	}

	<-daemon.Daemon.Stopped()
	<-server.Server.Stopped()
	if mqtt.MqttURL != "" {
		<-mqtt.MqttInstance.Stopped()
	}
	logger.Info("Runner terminated")
	os.Exit(0)

}
