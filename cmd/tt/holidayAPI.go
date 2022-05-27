// tt - an API tool for Trash Taxi
//
// Copyright 2018-2022 F5 Inc.
// Licensed under the BSD 3-clause license; see LICENSE.md for more information.

package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/urfave/cli"
)

type getHolidayResponse struct {
	Types []string
	Roles []string
}

type postResponse struct {
	Accepted bool
	Context  string
}

func getHolidays(c *cli.Context) {
	holidays, err := getHolidayAPI()
	if err != nil {
		fmt.Printf("*** Unable to get holidays: %s", err)
		os.Exit(1)
	}

	fmt.Printf("Excluded Roles: %v\n", holidays.Roles)
	fmt.Printf("Excluded Types: %v\n", holidays.Types)
}

func getHolidayAPI() (getHolidayResponse, error) {
	holidays := getHolidayResponse{}
	endpoint := fmt.Sprintf("%s/holidays", apiEndpoint)
	resp, err := getFromEndpoint(endpoint)
	if err != nil {
		return holidays, err
	}
	err = json.Unmarshal(resp, &holidays)
	if err != nil {
		return holidays, err
	}
	return holidays, nil
}

func addHoliday(c *cli.Context) {
	jsonResponse := postResponse{}

	if c.Args().Get(0) != "role" && c.Args().Get(0) != "type" {
		fmt.Println("*** Second argument to add should be role or type")
	}
	endpoint := fmt.Sprintf("%s/holiday/%s/%s", apiEndpoint, c.Args().Get(0), c.Args().Get(1))
	resp, err := postToEndpoint(endpoint)
	if err != nil {
		fmt.Printf("*** Could not get from endpoint: %s\n", err)
		os.Exit(1)
	}
	err = json.Unmarshal(resp, &jsonResponse)
	if err != nil {
		fmt.Printf("*** Could not unmarshal JSON: %s", err)
		os.Exit(1)
	}
	var reqState string
	if !jsonResponse.Accepted {
		reqState = "Request Denied"
	}

	fmt.Printf("%s", reqState)
	if jsonResponse.Context != "" {
		fmt.Printf(" (Server Message: %s)\n", jsonResponse.Context)
	}

	holidays, err := getHolidayAPI()
	if err != nil {
		fmt.Printf("*** Unable to get holidays: %s", err)
		os.Exit(1)
	}

	fmt.Printf("Excluded Roles: %v\n", holidays.Roles)
	fmt.Printf("Excluded Types: %v\n", holidays.Types)
}

func rmHoliday(c *cli.Context) {
	jsonResponse := postResponse{}
	if c.Args().Get(0) != "role" && c.Args().Get(0) != "type" {
		fmt.Println("*** Second argument to add should be role or type")
	}
	endpoint := fmt.Sprintf("%s/holiday/%s/%s", apiEndpoint, c.Args().Get(0), c.Args().Get(1))
	resp, err := deleteFromEndpoint(endpoint)
	if err != nil {
		fmt.Printf("*** Could not get from endpoint: %s\n", err)
		os.Exit(1)
	}
	err = json.Unmarshal(resp, &jsonResponse)
	if err != nil {
		fmt.Printf("*** Could not unmarshal JSON: %s", err)
		os.Exit(1)
	}
	var reqState string
	if !jsonResponse.Accepted {
		reqState = "Request Denied "
	}

	fmt.Printf("%s", reqState)
	if jsonResponse.Context != "Deleted role if it existed" {
		fmt.Printf("(Server Message: %s)\n", jsonResponse.Context)
	}

	holidays, err := getHolidayAPI()
	if err != nil {
		fmt.Printf("*** Unable to get holidays: %s", err)
		os.Exit(1)
	}
	fmt.Printf("Excluded Roles: %v\n", holidays.Roles)
	fmt.Printf("Excluded Types: %v\n", holidays.Types)

}
