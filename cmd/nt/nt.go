package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/user"
	"strings"
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
	config := NewConfig("/etc/nt.json")
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
	if currentUser.Uid == "0" || config.Local == false {
		getPKCS7, _ := http.Get(
			"http://169.254.169.254/latest/dynamic/instance-identity/pkcs7")
		defer getPKCS7.Body.Close()
		getPKCS7Body, _ := ioutil.ReadAll(getPKCS7.Body)
		pkcs7raw := strings.Replace(string(getPKCS7Body), "\n", "", -1)

		// send to endpoint
		endpoint := config.Endpoint + "/v1/trash/new"
		iid := map[string]string{"iid": pkcs7raw}
		shipIID, _ := json.Marshal(iid)
		resp, _ := http.Post(endpoint, "application/json", bytes.NewBuffer(shipIID))
		defer resp.Body.Close()

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
	proc, err := os.StartProcess(config.Shell, []string{config.Shell}, &pa)
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
