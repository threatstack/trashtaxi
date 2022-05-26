// garbaged - clean up manually modified hosts, quick
// db.go: Stuff related to databases - models and such
//
// Copyright 2018 Threat Stack, Inc.
// Licensed under the BSD 3-clause license; see LICENSE.md for more information.
// Author: Patrick T. Cable II <pat.cable@threatstack.com>

package server

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres" // Postgres driver
	"github.com/threatstack/trashtaxi/config"
)

func setupDB(dbcfg config.DatabaseConfig) (db *gorm.DB, err error) {
	if dbcfg.SSLMode == "" {
		dbcfg.SSLMode = "disable"
	}

	if dbcfg.SSLMode == "verify-full" {
		log.Info("Using postgres connection with TLS for database " + dbcfg.Name)
	} else {
		log.Info("Using postgres connection WITHOUT TLS for database " + dbcfg.Name)
	}

	var sslRootCertOption = ""
	if dbcfg.SSLRootCert != "" {
		sslRootCertOption = " sslrootcert=" + dbcfg.SSLRootCert
	}
	dbString := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s%s",
		dbcfg.Host, dbcfg.Port, dbcfg.User, dbcfg.Pass, dbcfg.Name, dbcfg.SSLMode, sslRootCertOption)
	db, err = gorm.Open("postgres", dbString)
	if conf.Debug {
		db.LogMode(true)
	} else {
		db.LogMode(false)
	}
	if err != nil {
		return nil, err
	}

	db.AutoMigrate(&Garbage{})
	db.AutoMigrate(&RoleHoliday{})
	db.AutoMigrate(&TypeHoliday{})

	return
}

// Garbage definition for ORM
type Garbage struct {
	gorm.Model
	Host    string `gorm:"UNIQUE"`
	Region  string
	Account string
	Role    string
	Type    string
}

// RoleHoliday definition
type RoleHoliday struct {
	gorm.Model
	Role string `gorm:"UNIQUE"`
	Conf bool
}

// TypeHoliday definition
type TypeHoliday struct {
	gorm.Model
	Type string `gorm:"UNIQUE"`
	Conf bool
}

// TableName definition
func (Garbage) TableName() string {
	return "trash"
}
