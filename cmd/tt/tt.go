package main

import (
	"log"
	"net/http"
	"os"

	"io/ioutil"

	"github.com/urfave/cli"
)

var apiEndpoint = "https://taxi.tls.zone/v1"

func main() {
	app := cli.NewApp()
	app.Name = "tt"
	app.Usage = "Send commands to the garbage daemon"
	app.Commands = []cli.Command{
		{
			Name:  "trash",
			Usage: "Display trash to be picked up",
			Action: func(c *cli.Context) error {
				getTrash(false, c)
				return nil
			},
			Subcommands: []cli.Command{
				{
					Name:  "all",
					Usage: "Display all trash",
					Action: func(c *cli.Context) error {
						getTrash(true, c)
						return nil
					},
				},
				{
					Name:  "pickup",
					Usage: "Pick up the trash (Terminate Instances!)",
					Action: func(c *cli.Context) error {
						pickupTrash(c)
						return nil
					},
				},
				{
					Name:  "cleanup",
					Usage: "Sort the trash (Remove stale nodes)",
					Action: func(c *cli.Context) error {
						cleanupTrash(c)
						return nil
					},
				},
			},
		},
		{
			Name:  "holiday",
			Usage: "Manage trash holidays",
			Action: func(c *cli.Context) error {
				getHolidays(c)
				return nil
			},
			Subcommands: []cli.Command{
				{
					Name:  "add",
					Usage: "Add a trash holiday",
					Action: func(c *cli.Context) error {
						addHoliday(c)
						return nil
					},
				},
				{
					Name:  "rm",
					Usage: "Remove a trash holiday",
					Action: func(c *cli.Context) error {
						rmHoliday(c)
						return nil
					},
				},
			},
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func getFromEndpoint(endpoint string) ([]byte, error) {
	res, err := http.Get(endpoint)
	if err != nil {
		return nil, err
	}
	json, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		return nil, readErr
	}
	return json, nil
}

func postToEndpoint(endpoint string) ([]byte, error) {
	res, err := http.Post(endpoint, "application/JSON", nil)
	if err != nil {
		return nil, err
	}
	json, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		return nil, readErr
	}
	return json, nil
}

func deleteFromEndpoint(endpoint string) ([]byte, error) {
	client := &http.Client{}
	req, err := http.NewRequest("DELETE", endpoint, nil)
	if err != nil {
		return nil, err
	}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	json, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		return nil, readErr
	}
	return json, nil
}
