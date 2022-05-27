// garbaged - clean up manually modified hosts, quick
// roothandler.go: The handler for /
//
// Copyright 2018-2022 F5 Inc.
// Licensed under the BSD 3-clause license; see LICENSE.md for more information.

package server

import (
	"fmt"
	"net/http"
)

// RootHandler responds to a GET to /
func rootHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Hop in and let's go to the landfill!\n")
}
