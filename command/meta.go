// garbaged - clean up manually modified hosts, quick
// meta.go: subcommand inheritance stuff
//
// Copyright 2018 Threat Stack, Inc.
// Licensed under the BSD 3-clause license; see LICENSE.md for more information.
// Author: Patrick T. Cable II <pat.cable@threatstack.com>

package command

import "github.com/mitchellh/cli"

// Meta contain the meta-option that nearly all subcommand inherits.
type Meta struct {
	UI cli.Ui
}
