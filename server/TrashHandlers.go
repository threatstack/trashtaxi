// garbaged - clean up manually modified hosts, quick
// TrashHandlers.go: The handler for all GET trash/ endpoints
//
// Copyright 2018 Threat Stack, Inc.
// Licensed under the BSD 3-clause license; see LICENSE.md for more information.
// Author: Patrick T. Cable II <pat.cable@threatstack.com>

package server

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// ListTrash - JSON output of hosts to be put into the rubbish bin
func listTrash(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var allTrash []Garbage
	var sortedTrash []Garbage
	types, roles := getHolidays()

	if err := db.Select("DISTINCT ON (role) *").Where("role NOT IN (?) AND type NOT IN (?)", roles, types).Limit(conf.AWS.TermLimit).Find(&allTrash).Error; err != nil {
		fmt.Printf("no: %v\n", err)
	}

	// go through the trash and pick one of each role
	sortedTrash = sortGarbage(allTrash, true, true)

	jsonOutput, _ := json.Marshal(&sortedTrash)
	w.Write(jsonOutput)
}

// ListAllTrash - JSON output of all hosts marked with a garbage flag,
// even those who are excluded from pickup
func listAllTrash(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var allTrash []Garbage
	var sortedTrash []Garbage

	if err := db.Find(&allTrash).Error; err != nil {
		fmt.Printf("no: %v\n", err)
	}

	// go through the trash and pick one of each role
	sortedTrash = sortGarbage(allTrash, false, false)

	jsonOutput, _ := json.Marshal(&sortedTrash)
	w.Write(jsonOutput)
}

// SortGarbage takes garbage and whether we're returning all of it
// and returns just that.
func sortGarbage(allTrash []Garbage, pickupRun bool, holidays bool) []Garbage {
	var seenRoles []string
	var sortedTrash []Garbage

	types, roles := getHolidays()

	for _, v := range allTrash {
		if pickupRun == true {
			if stringInSlice(v.Role, seenRoles) == true {
				continue
			}

			if stringInSlice(v.Role, roles) == true {
				continue
			}

			if stringInSlice(v.Type, types) == true {
				continue
			}
			seenRoles = append(seenRoles, v.Role)
		}
		if holidays == true {
			if stringInSlice(v.Role, roles) == true {
				continue
			}
			if stringInSlice(v.Type, types) == true {
				continue
			}
		}
		sortedTrash = append(sortedTrash, v)
	}
	return sortedTrash
}
