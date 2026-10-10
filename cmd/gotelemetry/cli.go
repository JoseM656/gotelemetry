package main

import (
	"flag"
)

type CLIArgs struct {
	ShowVersion bool
	RegenConfig bool
}

// ParseFlags procesa los argumentos de la línea de comandos
func ParseFlags() CLIArgs {
	var args CLIArgs

	flag.BoolVar(&args.ShowVersion, "version", false, "Shows version of the build")
	flag.BoolVar(&args.RegenConfig, "regen", false, "Regenerates a new default configuration file")

	flag.Parse()

	return args
}
