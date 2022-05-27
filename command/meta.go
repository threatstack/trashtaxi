// garbaged - clean up manually modified hosts, quick
// meta.go: subcommand inheritance header
//
// Copyright 2018-2022 F5 Inc.
// Licensed under the BSD 3-clause license; see LICENSE.md for more information.

package command

import "github.com/mitchellh/cli"

// Meta contain the meta-option that nearly all subcommand inherits.
type Meta struct {
	UI cli.Ui
}
