// nt - the tool to launch a new root shell
//
// Copyright 2018-2022 F5 Inc.
// Licensed under the BSD 3-clause license; see LICENSE.md for more information.

package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/user"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
)

// NTConfig - Configuration Struct for the app
type NTConfig struct {
	Shell    string `json:"shell"`
	Endpoint string `json:"endpoint"`
	Local    bool   `json:"local"`
}

// NewConfig - Pull in a new config from file
func NewConfig(fname string) NTConfig {
	data, err := ioutil.ReadFile(fname)
	if err != nil {
		fmt.Printf("/!\\ nt: Unable to read config at %s\n", fname)
		fmt.Printf("/!\\ nt: Does the config exist? Is it readable?\n")
		os.Exit(1)
	}
	config := NTConfig{}
	err = json.Unmarshal(data, &config)
	if err != nil {
		fmt.Printf("/!\\ nt: Unable to parse JSON. Is %s valid JSON?\n", fname)
		os.Exit(1)
	}
	return config
}

func main() {
	version := "1.0.0"
	ttConfig := NewConfig("/etc/nt.json")
	fmt.Println("/!\\/!\\/!\\/!\\/!\\/!\\/!\\/!\\/!\\/!\\/!\\/!\\/!\\/!\\/!\\/!\\/!\\/!\\/!\\/!\\/!\\/!\\/!\\")
	fmt.Println("nt marks hosts for later deletion! This is a LAST RESORT root shell!")
	fmt.Println(" Consider opening a ticket to fix this problem for future you!")
	fmt.Println("/!\\/!\\/!\\/!\\/!\\/!\\/!\\/!\\/!\\/!\\/!\\/!\\/!\\/!\\/!\\/!\\/!\\/!\\/!\\/!\\/!\\/!\\/!\\")
	fmt.Printf("Type 'yes' to mark host for later deletion & spawn a root shell: ")
	garbage := bufio.NewReader(os.Stdin)
	text, _ := garbage.ReadString('\n')
	text = strings.TrimSuffix(text, "\n")
	if text != "yes" {
		fmt.Println("<< nt: Didn't get 'yes' so I'm exiting.")
		os.Exit(1)
	}
	// Let's only break the seal if we're root or if local mode is set
	// get PKCS7 signed aws identity document
	currentUser, _ := user.Current()
	if currentUser.Uid == "0" || !ttConfig.Local {
		iid, sig, err := getInstanceIdentity()
		if err != nil {
			fmt.Println("<< nt: unable to get instance identity and signature")
		}

		// send to endpoint
		endpoint := ttConfig.Endpoint + "/v1/trash/new"
		payload := map[string]string{
			"iid":       sig,
			"signature": iid,
		}
		shipIID, _ := json.Marshal(payload)
		resp, err := http.Post(endpoint, "application/json", bytes.NewBuffer(shipIID))
		if err != nil {
			fmt.Printf("<< nt: unable to post response: %s", err.Error())
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			fmt.Printf("<< nt: Unable to register host, server responded with %d.\n", resp.StatusCode)
			os.Exit(1)
		}

		fmt.Printf(">> EPA Registration: %s\n", resp.Status)
	}

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Println("<< nt: Unable to get current working directory (?!)")
		os.Exit(1)
	}
	pa := os.ProcAttr{
		Files: []*os.File{os.Stdin, os.Stdout, os.Stderr},
		Dir:   cwd,
	}

	fmt.Printf(">> nt: begin session (v%s)\n", version)
	proc, err := os.StartProcess(ttConfig.Shell, []string{ttConfig.Shell}, &pa)
	if err != nil {
		fmt.Println("<< nt: Unable to reassign stdin/stdout/stderr (?!)")
		os.Exit(1)
	}

	state, err := proc.Wait()
	if err != nil {
		fmt.Println("<< nt: Yikes! Something bad happened in the wait loop.")
		os.Exit(1)
	}

	fmt.Printf("<< nt: thanks for coming in! (%s)\n", state.String())
}

func getInstanceIdentity() (string, string, error) {
	ctx := context.TODO()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return "", "", err
	}

	client := imds.NewFromConfig(cfg)
	iidBytes, err := client.GetDynamicData(ctx, &imds.GetDynamicDataInput{
		Path: "instance-identity/document",
	})
	if err != nil {
		return "", "", err
	}

	iidBuf := new(strings.Builder)
	_, err = io.Copy(iidBuf, iidBytes.Content)
	if err != nil {
		return "", "", err
	}

	iid := base64.StdEncoding.EncodeToString([]byte(iidBuf.String()))

	fmt.Printf("%+v\n", iidBuf.String())

	sigBytes, err := client.GetDynamicData(ctx, &imds.GetDynamicDataInput{
		Path: "instance-identity/signature",
	})
	if err != nil {
		return "", "", err
	}

	sigBuf := new(strings.Builder)
	_, err = io.Copy(sigBuf, sigBytes.Content)
	if err != nil {
		return "", "", err
	}

	sig := base64.StdEncoding.EncodeToString([]byte(sigBuf.String()))

	return iid, sig, nil
}
