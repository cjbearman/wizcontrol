package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/cjbearman/framework/pkg/logger"
	"github.com/cjbearman/wizcontrol2/pkg/wizcontrol/common"
	"github.com/cjbearman/wizcontrol2/pkg/wizcontrol/config"
)

func main() {
	logLevel := flag.String("log", "info", "Set the logging level")
	cfg := flag.String("config", "config.yaml", "Path to the configuration file")
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

}
