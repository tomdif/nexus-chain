package main

import (
	"os"

	"nexus/cmd/nexusd/cmd"
)

func main() {
	rootCmd := cmd.NewRootCmd()
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
