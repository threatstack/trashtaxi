// lambda-invoker - a sample lambda invoker for trash taxi
//
// Copyright 2018-2022 F5 Inc.
// Licensed under the BSD 3-clause license; see LICENSE.md for more information.

package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/aws/aws-lambda-go/lambda"
)

var apiEndpoint = "https://taxi.tls.zone/v1"

var rootca = `-----BEGIN CERTIFICATE-----
MIICzTCCAlOgAwIBAgIURrXSQ94pBpE/hGr8on9+bUu2AuUwCgYIKoZIzj0EAwMw
gYgxCzAJBgNVBAYTAlVTMRYwFAYDVQQIEw1NYXNzYWNodXNldHRzMQ8wDQYDVQQH
EwZCb3N0b24xGzAZBgNVBAoTElRocmVhdCBTdGFjaywgSW5jLjEUMBIGA1UECxML
RW5naW5lZXJpbmcxHTAbBgNVBAMTFFRocmVhdCBTdGFjayBSb290IENBMB4XDTE2
MTAxNzE3NTAwMFoXDTI2MTAxNTE3NTAwMFowgYgxCzAJBgNVBAYTAlVTMRYwFAYD
VQQIEw1NYXNzYWNodXNldHRzMQ8wDQYDVQQHEwZCb3N0b24xGzAZBgNVBAoTElRo
cmVhdCBTdGFjaywgSW5jLjEUMBIGA1UECxMLRW5naW5lZXJpbmcxHTAbBgNVBAMT
FFRocmVhdCBTdGFjayBSb290IENBMHYwEAYHKoZIzj0CAQYFK4EEACIDYgAEfKSB
llVLhHJle/6DyeKdvNrJEMrh3/vhbRZEv7CGiN5QHG1r6QljnjrOXctWhzURlXrj
cEdHZEv3VOksRPRRFac4LX32uVqcIhzLPaXBjjQjNebE1S4LI80d9t5uK/Teo3ww
ejAOBgNVHQ8BAf8EBAMCAQYwEgYDVR0TAQH/BAgwBgEB/wIBATAdBgNVHQ4EFgQU
R5h8ktJHUJJDpwI3/MMLgIKsICswNQYDVR0fBC4wLDAqoCigJoYkaHR0cHM6Ly9j
cmwudGhyZWF0c3RhY2submV0L3RzY2EuY3JsMAoGCCqGSM49BAMDA2gAMGUCMQCX
3Y7pgeAFDopHyySUc1RBZ1LW9NuBIjw5Bbc9ggE3ukAVsrqR+pPRK1Xjs6PXXcAC
MHqbYjArUdowgMkWZeBDfs0eDGzKI0iAcm2GI/WdfWGwPEYB6kT0QZQ34liDF6fb
Lg==
-----END CERTIFICATE-----`

type postResponse struct {
	Accepted bool
	Context  string
}

func doStuff(ctx context.Context) (string, error) {
	endpoint := fmt.Sprintf("%s/trash/pickup", apiEndpoint)
	jsonResponse := postResponse{}

	// Set up TLS stuff
	roots := x509.NewCertPool()
	if ok := roots.AppendCertsFromPEM([]byte(rootca)); !ok {
		log.Println("No certs appended, using system certs only")
	}
	config := &tls.Config{RootCAs: roots}
	tr := &http.Transport{TLSClientConfig: config}
	client := &http.Client{Transport: tr}

	req, err := http.NewRequest(http.MethodPost, endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("*** Could not set up request: %s", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("*** Could not POST to endpoint: %s", err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("*** Could not decode body: %s", err)
	}

	err = json.Unmarshal(body, &jsonResponse)
	if err != nil {
		return "", fmt.Errorf("*** Could not unmarshal JSON: %s", err)
	}

	var reqState string
	if jsonResponse.Accepted {
		reqState = "Request Accepted"
	} else {
		reqState = "Request Denied"
	}

	if jsonResponse.Context != "" {
		reqState = reqState + fmt.Sprintf(" (Server Message: %s)", jsonResponse.Context)
	}

	return reqState, nil
}

func main() {
	lambda.Start(doStuff)
}
