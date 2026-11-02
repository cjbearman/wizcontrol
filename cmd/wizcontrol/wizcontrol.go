package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/cjbearman/framework/pkg/logger"
	"github.com/cjbearman/wizcontrol2/pkg/wizcontrol/calendar"
	"github.com/cjbearman/wizcontrol2/pkg/wizcontrol/common"
	"github.com/cjbearman/wizcontrol2/pkg/wizcontrol/config"
	"github.com/cjbearman/wizcontrol2/pkg/wizcontrol/daemon"
	"github.com/cjbearman/wizcontrol2/pkg/wizcontrol/operations"
	"github.com/cjbearman/wizcontrol2/pkg/wizcontrol/scenes"
)

func main() {
	logLevel := flag.String("log", "info", "Set the logging level")
	cfg := flag.String("config", "config.yaml", "Path to the configuration file")
	apply := flag.String("apply", "", "Apply scene")
	f := flag.Bool("f", false, "Force apply scene")
	d := flag.Bool("d", false, "Run as daemon")
	cd := flag.Bool("cd", false, "Enable calendar debug")

	flag.Parse()

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

	common.Logger.Infof("Operating in location: %v", common.Location)

	if *apply != "" {

		_, found := scenes.Scenes[*apply]
		if !found {
			common.Logger.Errorf("Scene %q not found", *apply)
			os.Exit(1)
		}
		operations.UpdateAvailability(true)
		operations.ApplyScenes([]string{*apply}, *f)
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

	// Start the main daemon
	daemon.Daemon.Start()

	select {

	case <-daemon.Daemon.Stopped():
		logger.Warn("Daemon has stopped")

	case sig := <-sigch:
		logger.Warnf("Caught signal %v, stopping daemon", sig)
		daemon.Daemon.Stop()
	}

	<-daemon.Daemon.Stopped()
	logger.Info("Runner terminated")
	os.Exit(0)
}
