// garbaged - clean up manually modified hosts, quick
// HolidayHandler.go - Handlers for Trash Holidays
//
// Copyright 2018 Threat Stack, Inc.
// Licensed under the BSD 3-clause license; see LICENSE.md for more information.
// Author: Patrick T. Cable II <pat.cable@threatstack.com>

package server

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

type holidayIO struct {
	Types []string `json:"types"`
	Roles []string `json:"roles"`
}

// ListHolidays: Endpoint to list all the holidays.
func listHolidays(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	types, roles := getHolidays()
	holidays := holidayIO{Types: types, Roles: roles}
	jsonOutput, _ := json.Marshal(&holidays)
	w.Write(jsonOutput)
}

// NewRoleHoliday: Schedule a Trash Holiday for a role.
func newRoleHoliday(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	myRoleToAdd := strconv.QuoteToASCII(params["name"])
	myRoleToAdd = myRoleToAdd[1 : len(myRoleToAdd)-1]
	log.Infof("/v1/holiday/role/%s: POST %s", myRoleToAdd, r.RemoteAddr)

	data := RoleHoliday{Role: myRoleToAdd, Conf: false}
	if err := db.Create(&data).Error; err != nil {
		resp := response{Accepted: false, Context: err.Error()}
		jsonOutput, _ := json.Marshal(&resp)
		w.Write(jsonOutput)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	resp := response{Accepted: true}
	jsonOutput, _ := json.Marshal(&resp)
	w.Write(jsonOutput)
}

// NewTypeHoliday: Schedule a Trash Holiday for a role
func newTypeHoliday(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	myTypeToAdd := strconv.QuoteToASCII(params["name"])
	myTypeToAdd = myTypeToAdd[1 : len(myTypeToAdd)-1]
	log.Infof("/v1/holiday/type/%s: POST %s", myTypeToAdd, r.RemoteAddr)

	data := TypeHoliday{Type: myTypeToAdd, Conf: false}
	if err := db.Create(&data).Error; err != nil {
		resp := response{
			Accepted: false,
			Context:  err.Error(),
		}
		jsonOutput, _ := json.Marshal(&resp)
		w.Write(jsonOutput)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	resp := response{Accepted: true}
	jsonOutput, _ := json.Marshal(&resp)
	w.Write(jsonOutput)
}

// DeleteRoleHoliday: Deletes a role holiday
func deleteRoleHoliday(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	myRoleToDelete := strconv.QuoteToASCII(params["name"])
	myRoleToDelete = myRoleToDelete[1 : len(myRoleToDelete)-1]
	log.Infof("/v1/holiday/role/%s: DELETE %s", myRoleToDelete, r.RemoteAddr)

	// Dont delete roles that are in the stateful config.
	if stringInSlice(myRoleToDelete, conf.Stateful.Roles) {
		resp := response{
			Accepted: false,
			Context:  "Role is defined as stateful in config file",
		}
		jsonOutput, _ := json.Marshal(&resp)
		w.Write(jsonOutput)
		return
	}

	// Actually delete the data. We use unscoped() because we want gorm to
	// fully delete the data, vs. set the DeletedAt field.
	data := RoleHoliday{Role: myRoleToDelete, Conf: false}
	if err := db.Where("role LIKE ?", myRoleToDelete).Unscoped().Delete(&data).Error; err != nil {
		resp := response{
			Accepted: false,
			Context:  err.Error(),
		}
		jsonOutput, _ := json.Marshal(&resp)
		w.Write(jsonOutput)
		return
	}

	// Let the user know
	w.Header().Set("Content-Type", "application/json")
	resp := response{
		Accepted: true,
		Context:  "Deleted role if it existed",
	}
	jsonOutput, _ := json.Marshal(&resp)
	w.Write(jsonOutput)
}

// DeleteTypeHoliday: Deletes a type holiday.
func deleteTypeHoliday(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	myTypeToDelete := strconv.QuoteToASCII(params["name"])
	myTypeToDelete = myTypeToDelete[1 : len(myTypeToDelete)-1]
	log.Infof("/v1/holiday/type/%s: DELETE %s", myTypeToDelete, r.RemoteAddr)

	// Dont delete types that are in the stateful config.
	if stringInSlice(myTypeToDelete, conf.Stateful.Types) {
		resp := response{
			Accepted: false,
			Context:  "Type is defined as stateful in config file",
		}
		jsonOutput, _ := json.Marshal(&resp)
		w.Write(jsonOutput)
		return
	}

	// Actually delete the data. We use unscoped() because we want gorm to
	// fully delete the data, vs. set the DeletedAt field.
	data := TypeHoliday{Type: myTypeToDelete, Conf: false}
	if err := db.Where("type LIKE ?", myTypeToDelete).Unscoped().Delete(&data).Error; err != nil {
		resp := response{
			Accepted: false,
			Context:  err.Error(),
		}
		jsonOutput, _ := json.Marshal(&resp)
		w.Write(jsonOutput)
		return
	}

	// Let the user know
	w.Header().Set("Content-Type", "application/json")
	resp := response{
		Accepted: true,
		Context:  "Deleted type if it existed",
	}
	jsonOutput, _ := json.Marshal(&resp)
	w.Write(jsonOutput)
}

// getHolidays: return all type/role holidays. An internal function for the
// server to use in a variety of places.
func getHolidays() (holidayTypes []string, holidayRoles []string) {
	var roleRecords []RoleHoliday
	var typeRecords []TypeHoliday

	if err := db.Find(&roleRecords).Error; err != nil {
		log.Warnf("Unable to get role holidays: %s\n", err)
	}
	if err := db.Find(&typeRecords).Error; err != nil {
		log.Warnf("Unable to get type holidays: %s\n", err)
	}

	for _, v := range typeRecords {
		holidayTypes = append(holidayTypes, v.Type)
	}
	for _, v := range roleRecords {
		holidayRoles = append(holidayRoles, v.Role)
	}

	return
}
