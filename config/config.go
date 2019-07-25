// garbaged - clean up manually modified hosts, quick
// config.go: set up a config file & parse it
//
// Copyright 2018 Threat Stack, Inc.
// Licensed under the BSD 3-clause license; see LICENSE.md for more information.
// Author: Patrick T. Cable II <pat.cable@threatstack.com>

package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

// Config is the global config object
var Config GarbagedConfig

// ConfigFile is the path to a config file defined by GARBAGED_CONFIG env
var ConfigFile string

func init() {
	if os.Getenv("GARBAGED_CONFIG") == "" {
		ConfigFile = "/etc/garbaged.json"
	} else {
		ConfigFile = os.Getenv("GARBAGED_CONFIG")
	}
	if _, err := os.Stat(ConfigFile); err == nil {
		// File exists!
		if Config, err = NewConfig(ConfigFile); err != nil {
			// Newconfig passed error
			fmt.Printf("%v", err)
			os.Exit(1)
		}
	}
}

// DatabaseConfig holds info on the database.
type DatabaseConfig struct {
	Host        string `json:"host"`
	Port        int    `json:"port"`
	User        string `json:"user"`
	Pass        string `json:"pass"`
	Name        string `json:"name"`
	SSLMode     string `json:"sslmode"`
	SSLRootCert string `json:"sslrootcert"`
}

// AWSAccount holds info about AWS accounts. Indexed by account number.
type AWSAccount struct {
	ARN        string `json:"arn"`
	ExternalID string `json:"externalid"`
	Name       string `json:"name"`
}

// AWSConfig is AWS-Specific config bits.
type AWSConfig struct {
	RoleEC2Tag string `json:"role_ec2_tag"`
	TypeEC2Tag string `json:"type_ec2_tag"`
	DryRun     bool   `json:"dryrun"`
	TermLimit  int    `json:"termlimit"`
}

// StatefulConfig defines which node types are permanently stateful
type StatefulConfig struct {
	Roles []string `json:"roles"`
	Types []string `json:"types"`
}

// GarbagedConfig is our config struct
type GarbagedConfig struct {
	Debug             bool                  `json:"debug"`
	Bind              string                `json:"bind"`
	TLSKey            string                `json:"tlskey"`
	TLSCert           string                `json:"tlscert"`
	IIDCert           string                `json:"iid_verify_cert"`
	Database          DatabaseConfig        `json:"database"`
	Accounts          map[string]AWSAccount `json:"accounts"`
	AWS               AWSConfig             `json:"aws"`
	Stateful          StatefulConfig        `json:"stateful"`
	Graylog           GraylogConfig         `json:"graylog"`
	TimeBetweenPickup string                `json:"timebetweenpickup"`
	Slack             SlackConfig           `json:"slack"`
}

// GraylogConfig configs... graylog
type GraylogConfig struct {
	Enabled bool   `json:"enabled"`
	Host    string `json:"host"`
	Port    int    `json:"port"`
}

// SlackConfig - for your slack config needs
type SlackConfig struct {
	Enabled  bool     `json:"enabled"`
	APIKey   string   `json:"APIKey"`
	Channels []string `json:"channels"`
	TimeWait string   `json:"timewait"`
}

// NewConfig reads in a config file and set config
func NewConfig(fname string) (config GarbagedConfig, err error) {
	data, err := ioutil.ReadFile(fname)
	if err != nil {
		return config, fmt.Errorf("!!! could not read %s (%s)", fname, err)
	}

	config = GarbagedConfig{}
	err = json.Unmarshal(data, &config)
	if err != nil {
		return config, fmt.Errorf("!!! could not deserialize %s (%s)", fname, err)
	}
	return config, nil
}
