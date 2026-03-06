// Package main is the entry point for md-paste.
package main

import (
	"fmt"
	"os"

	"github.com/stn1slv/md-paste/internal/cli"
	"github.com/stn1slv/md-paste/internal/logger"
)

const exitFailure = 1

func main() {
	logger.Init()
	if err := cli.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(exitFailure)
	}
}
