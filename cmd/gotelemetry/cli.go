package main

import (
	"flag"
)

type CLIArgs struct {
	// ConfigPathFlag string
	ShowVersion bool
}

// ParseFlags procesa los argumentos de la línea de comandos
func ParseFlags() CLIArgs {
	var args CLIArgs

	// flag.StringVar(&args.ConfigPathFlag, "config", "", "Path to configuration file")
	// flag.StringVar(&args.ConfigPathFlag, "c", "", "Path to configuration file (abbreviated)")
	flag.BoolVar(&args.ShowVersion, "version", false, "Shows version of the build")

	flag.Parse()

	return args
}
