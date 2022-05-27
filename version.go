// garbaged - clean up manually modified hosts, quick
// version.go: CLI versioning
//
// Copyright 2018-2022 F5 Inc.
// Licensed under the BSD 3-clause license; see LICENSE.md for more information.

package main

// Name of the application
const Name string = "garbaged"

// Version is the current version
const Version string = "1.1.0"

// GitCommit describes latest commit hash.
// This value is extracted by git command when building.
// To set this from outside, use go build -ldflags "-X main.GitCommit \"$(COMMIT)\""
var GitCommit string
