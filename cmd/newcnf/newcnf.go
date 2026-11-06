package main

import (
	"flag"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/newconfig"
)

func main() {
	config := flag.String("config", "newconfig.yaml", "Path to the new configuration YAML file")
	flag.Parse()

	data, err := newconfig.LoadNew(*config)
	if err != nil {
		panic(err)
	}

	if err := newconfig.Build(data); err != nil {
		panic(err)
	}
}
