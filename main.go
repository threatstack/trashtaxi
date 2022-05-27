// garbaged - clean up manually modified hosts, quick
// main.go: CLI initialization
//
// Copyright 2018-2022 F5 Inc.
// Licensed under the BSD 3-clause license; see LICENSE.md for more information.

package main

import "os"

func main() {
	os.Exit(Run(os.Args[1:]))
}
