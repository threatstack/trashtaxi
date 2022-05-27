// garbaged - clean up manually modified hosts, quick
// server.go: struct for running the garbaged server
//
// Copyright 2018-2022 F5 Inc.
// Licensed under the BSD 3-clause license; see LICENSE.md for more information.

package command

import (
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/threatstack/trashtaxi/config"
	"github.com/threatstack/trashtaxi/server"
)

// ServerCommand gets the data from the CLI
type ServerCommand struct {
	Meta
}

// Run actually.. does the stuff.
func (c *ServerCommand) Run(args []string) int {
	if _, err := os.Stat(config.ConfigFile); os.IsNotExist(err) {
		log.Errorf("No config file present. See README.md.\n")
		os.Exit(1)
	}

	var conf = config.Config

	if conf.Debug == true {
		server.PrintConfig(conf)
	}

	server.Start(conf)

	return 0
}

// Synopsis gives the help output for server
func (c *ServerCommand) Synopsis() string {
	return "Starts the garbaged server"
}

// Help prints more useful info for server
func (c *ServerCommand) Help() string {
	helpText := `
Usage: garbaged server

  This command starts the garbaged server.
`
	return strings.TrimSpace(helpText)
}
