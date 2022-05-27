// tt - an API tool for Trash Taxi
//
// Copyright 2018-2022 F5 Inc.
// Licensed under the BSD 3-clause license; see LICENSE.md for more information.

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/urfave/cli"
)

type trashResponse []trash

type trash struct {
	CreatedAt string
	Host      string
	Region    string
	Account   string
	Role      string
	Type      string
}

func getTrash(all bool, c *cli.Context) {
	trash := trashResponse{}
	var endpoint string
	if !all {
		endpoint = fmt.Sprintf("%s/trash", apiEndpoint)
	} else {
		endpoint = fmt.Sprintf("%s/trash/all", apiEndpoint)
	}
	resp, err := getFromEndpoint(endpoint)
	if err != nil {
		fmt.Printf("*** Could not get from endpoint: %s\n", err)
		os.Exit(1)
	}
	err = json.Unmarshal(resp, &trash)
	if err != nil {
		fmt.Printf("*** Could not unmarshal JSON: %s", err)
		os.Exit(1)
	}

	for _, v := range trash {
		fmt.Printf("[%s] %s/%s in %s (%s/%s)\n", v.CreatedAt, v.Host, v.Account, v.Region, v.Role, v.Type)
	}
}

type cleanupResponse struct {
	Accepted bool
	Nodes    map[string]map[string][]string
	Context  string
}

func cleanupTrash(c *cli.Context) {
	endpointResponse := cleanupResponse{}
	endpoint := fmt.Sprintf("%s/trash/cleanup", apiEndpoint)
	resp, err := postToEndpoint(endpoint)
	if err != nil {
		fmt.Printf("*** Could not get from endpoint: %s\n", err)
		os.Exit(1)
	}
	err = json.Unmarshal(resp, &endpointResponse)
	if err != nil {
		fmt.Printf("*** Could not unmarshal JSON: %s", err)
		os.Exit(1)
	}
	for region, accounts := range endpointResponse.Nodes {
		for account, instances := range accounts {
			if instances != nil {
				fmt.Printf("%s/%s removed %v\n", region, account, instances)
			}
		}
	}
}

func pickupTrash(c *cli.Context) {
	if c.Args().Get(0) != "noninteractive" {
		fmt.Println("/!\\/!\\/!\\/!\\/!\\/!\\/!\\/!\\/!\\/!\\/!\\/!\\/!\\/!\\/!\\/!\\/!\\/!\\/!\\/!\\/!\\/!\\/!\\")
		fmt.Println("THIS WILL TERMINATE SERVERS. VALIDATE THE tt trash OUTPUT BEFORE CONTINUING.")
		fmt.Println("/!\\/!\\/!\\/!\\/!\\/!\\/!\\/!\\/!\\/!\\/!\\/!\\/!\\/!\\/!\\/!\\/!\\/!\\/!\\/!\\/!\\/!\\/!\\")
		fmt.Printf("Type yes to continue: ")
		garbage := bufio.NewReader(os.Stdin)
		text, _ := garbage.ReadString('\n')
		text = strings.TrimSuffix(text, "\n")
		if text != "yes" {
			fmt.Println("<< tt: Didn't get 'yes' so I'm exiting.")
			os.Exit(1)
		}
	}
	jsonResponse := postResponse{}
	endpoint := fmt.Sprintf("%s/trash/pickup", apiEndpoint)
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
	if jsonResponse.Accepted {
		reqState = "Request Accepted"
	} else {
		reqState = "Request Denied"
	}

	fmt.Printf("%s", reqState)
	if jsonResponse.Context != "" {
		fmt.Printf(" (Server Message: %s)", jsonResponse.Context)
	}
	fmt.Printf("\n")
}
