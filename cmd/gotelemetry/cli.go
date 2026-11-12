package main

import (
	"github.com/spf13/cobra"
)

/*
Cobra cli:
	Commands:
		version
		regen
		path

	flags
		--path
*/

type Options struct {
	Command    string
	ConfigPath string
}

func ParseFlags() Options {
	opts := Options{}

	// Comando raíz, osea cuando no se elige ninguna opcion: "gotelemetry"
	var rootCmd = &cobra.Command{
		Use:   "gotelemetry",
		Short: "Monolithic daemon for telemetry and power consumption monitoring",
		Run: func(cmd *cobra.Command, args []string) {

			opts.Command = "daemon"
		},
	}

	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Shows gotelemetry version.",
		Run: func(cmd *cobra.Command, args []string) {
			opts.Command = "version"
		},
	}

	var regenCmd = &cobra.Command{
		Use:   "regen",
		Short: "Regenerates a new default configuration file in the available path",
		Run: func(cmd *cobra.Command, args []string) {
			opts.Command = "regen"
		},
	}

	var pathCmd = &cobra.Command{
		Use:   "path",
		Short: "Check and display which configuration path is active",
		Run: func(cmd *cobra.Command, args []string) {
			opts.Command = "path"
		},
	}

	// Flag global --path (-p) disponible para cualquier comando
	rootCmd.PersistentFlags().StringVarP(&opts.ConfigPath, "path", "p", "", "Custom path for configuration")

	// Registrar subcomandos en la raíz
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(regenCmd)
	rootCmd.AddCommand(pathCmd)

	// Ejecutar el parser de Cobra
	_ = rootCmd.Execute()

	return opts
}
