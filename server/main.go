// garbaged - clean up manually modified hosts, quick
// main.go: globals and server init
//
// Copyright 2018 Threat Stack, Inc.
// Licensed under the BSD 3-clause license; see LICENSE.md for more information.
// Author: Patrick T. Cable II <pat.cable@threatstack.com>

package server

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
	"github.com/threatstack/trashtaxi/config"
	graylog "gopkg.in/gemnasium/logrus-graylog-hook.v2"
)

// Global vars for server: conf and db and accounts
var db *gorm.DB
var conf = config.Config
var knownAWSAccts []string
var lastPickupTime time.Time
var pickupLock bool

// How do we report results? Responses.
type response struct {
	Accepted bool   `json:"accepted"`
	Context  string `json:"context"`
}

// PrintConfig just prints config and was used to make sure that structs
// were properly implemented.
func PrintConfig(conf config.GarbagedConfig) error {
	log.Debugf("Database config: host '%s' port '%d' user '%s' pass '*' tls '%s'\n",
		conf.Database.Host, conf.Database.Port, conf.Database.User,
		conf.Database.SSLMode)
	for acct, info := range conf.Accounts {
		log.Debugf("Account %s - Name %s, ARN %s\n", acct, info.Name, info.ARN)
	}
	log.Debugf("AWS Options: Role tag is \"%s\", Type tag is \"%s\"\n",
		conf.AWS.RoleEC2Tag, conf.AWS.TypeEC2Tag)
	log.Debugf("Stateful roles: %s\n", conf.Stateful.Roles)
	log.Debugf("Stateful types: %s\n", conf.Stateful.Types)

	return nil
}

// Start will bring our server up and start listening. Hooray.
func Start(conf config.GarbagedConfig) error {
	var err error
	db, err = setupDB(conf.Database)
	if err != nil {
		log.Errorf("Unable to connect to database: %s\n", err)
		os.Exit(1)
	}

	for _, v := range conf.Stateful.Types {
		h := TypeHoliday{Type: v, Conf: true}
		if err := db.Create(&h).Error; err != nil {
			log.Warnf("Could not add Stateful Type (Continuing): %s\n", err.Error())
		}
	}

	for _, v := range conf.Stateful.Roles {
		h := RoleHoliday{Role: v, Conf: true}
		if err := db.Create(&h).Error; err != nil {
			log.Warnf("Could not add Stateful Role (Continuing): %s\n", err.Error())
		}
	}

	for k := range conf.Accounts {
		knownAWSAccts = append(knownAWSAccts, k)
	}

	if conf.Debug == true {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	if conf.Graylog.Enabled == true {
		connString := fmt.Sprintf("%s:%d", conf.Graylog.Host, conf.Graylog.Port)
		hook := graylog.NewAsyncGraylogHook(connString, map[string]interface{}{"facility": "garbaged"})
		defer hook.Flush()
		log.AddHook(hook)
	}
	log.Infof("Starting Garbaged (listening on %s)\n", conf.Bind)
	r := mux.NewRouter()
	r.HandleFunc("/", rootHandler).Methods("GET")
	r.HandleFunc("/v1/trash", listTrash).Methods("GET")
	r.HandleFunc("/v1/trash/all", listAllTrash).Methods("GET")
	r.HandleFunc("/v1/trash/new", newTrash).Methods("POST")
	r.HandleFunc("/v1/trash/pickup", pickupTrash).Methods("POST")
	r.HandleFunc("/v1/trash/cleanup", cleanupTrash).Methods("POST")
	r.HandleFunc("/v1/holidays", listHolidays).Methods("GET")
	r.HandleFunc("/v1/holiday/role/{name}", newRoleHoliday).Methods("POST")
	r.HandleFunc("/v1/holiday/type/{name}", newTypeHoliday).Methods("POST")
	r.HandleFunc("/v1/holiday/role/{name}", deleteRoleHoliday).Methods("DELETE")
	r.HandleFunc("/v1/holiday/type/{name}", deleteTypeHoliday).Methods("DELETE")
	err = http.ListenAndServeTLS(conf.Bind, conf.TLSCert, conf.TLSKey, r)
	if err != nil {
		log.Errorf("Could not start TLS server: %s\n", err)
	}

	return nil
}

func removeDuplicatesUnordered(elements []string) []string {
	encountered := map[string]bool{}

	// Create a map of all unique elements.
	for v := range elements {
		encountered[elements[v]] = true
	}

	// Place all keys from the map into a slice.
	result := []string{}
	for key := range encountered {
		result = append(result, key)
	}
	return result
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
