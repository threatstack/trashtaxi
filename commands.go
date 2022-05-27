// garbaged - clean up manually modified hosts, quick
// commands.go: CLI initialization
//
// Copyright 2018-2022 F5 Inc.
// Licensed under the BSD 3-clause license; see LICENSE.md for more information.

package main

import (
	"github.com/mitchellh/cli"
	"github.com/threatstack/trashtaxi/command"
)

// Commands is the factory generator for various command options in Deputize
func Commands(meta *command.Meta) map[string]cli.CommandFactory {
	return map[string]cli.CommandFactory{
		"server": func() (cli.Command, error) {
			return &command.ServerCommand{
				Meta: *meta,
			}, nil
		},

		"version": func() (cli.Command, error) {
			return &command.VersionCommand{
				Meta:     *meta,
				Version:  Version,
				Revision: GitCommit,
				Name:     Name,
			}, nil
		},
	}
}
