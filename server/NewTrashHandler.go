// garbaged - clean up manually modified hosts, quick
// NewTrashHandler.go: The handler for trash/new
//
// Copyright 2018-2022 F5 Inc.
// Licensed under the BSD 3-clause license; see LICENSE.md for more information.

package server

import (
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/service/ec2"
	log "github.com/sirupsen/logrus"
)

type incomingTrash struct {
	IID       string `json:"iid"`
	Signature string `json:"signature"`
}

// NewTrash - Add a new host
func newTrash(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var trash incomingTrash
	buffer, _ := ioutil.ReadAll(r.Body)
	log.Infof("/v1/trash/new: POST %s", r.RemoteAddr)
	log.Debugf("/v1/trash/new: POST %s (PAYLOAD: %v)",
		r.RemoteAddr, string(buffer))

	json.Unmarshal(buffer, &trash)

	rawIID, err := base64.StdEncoding.DecodeString(trash.IID)
	if err != nil {
		log.Warnf("/v1/trash/new: Unable to un-base64 IID")
		resp := response{Accepted: false, Context: "Unable to un-base64 IID"}
		jsonOutput, _ := json.Marshal(&resp)
		w.WriteHeader(http.StatusForbidden)
		w.Write(jsonOutput)
		return
	}

	var myIID ec2metadata.EC2InstanceIdentityDocument
	json.Unmarshal(rawIID, &myIID)

	// Perform verification.
	validate := verifyIID(rawIID, []byte(trash.Signature))
	if validate != nil {
		log.Warnf("/v1/trash/new: Unable to verify instance document")
		resp := response{Accepted: false, Context: "Unable to verify instance document"}
		jsonOutput, _ := json.Marshal(&resp)
		w.WriteHeader(http.StatusForbidden)
		w.Write(jsonOutput)
		return
	}

	// Do we know about this account? If not, let's not accept garbage from it.
	if !stringInSlice(myIID.AccountID, knownAWSAccts) {
		text := fmt.Sprintf("Cant collect hosts in unknown account (account %s not configured)", myIID.AccountID)
		log.Warnf("/v1/trash/new: %s", text)
		resp := response{Accepted: false, Context: text}
		jsonOutput, _ := json.Marshal(&resp)
		w.WriteHeader(http.StatusForbidden)
		w.Write(jsonOutput)
		return
	}

	// Auth using existing role credentials
	awsSession, err := awsSession(myIID.Region)
	if err != nil {
		text := fmt.Sprintf("Unable to start AWS session (%s)", err)
		log.Warnf("/v1/trash/new: %s", text)
		resp := response{Accepted: false, Context: text}
		jsonOutput, _ := json.Marshal(&resp)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(jsonOutput)
		return
	}

	// AssumeRole into the appropriate cross-account role
	creds := awsStsCreds(conf.Accounts[myIID.AccountID].ARN,
		conf.Accounts[myIID.AccountID].ExternalID, awsSession)
	// dont actually need creds output; just need to test for errors.
	if _, err := creds.Get(); err != nil {
		text := fmt.Sprintf("Unable to get AWS STS Credentials (%s)", err)
		log.Warnf("/v1/trash/new: %s", text)
		resp := response{Accepted: false, Context: text}
		jsonOutput, _ := json.Marshal(&resp)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(jsonOutput)
		return
	}

	// Start a new EC2 session and get values for Role/Type
	svc := ec2.New(awsSession, &aws.Config{Credentials: creds})
	ec2role, ec2type, err := getTags(svc, myIID.InstanceID,
		conf.AWS.RoleEC2Tag, conf.AWS.TypeEC2Tag)
	if err != nil {
		text := fmt.Sprintf("Unable to GetTags (%s)", err)
		log.Warnf("/v1/trash/new: %s", text)
		resp := response{Accepted: false, Context: text}
		jsonOutput, _ := json.Marshal(&resp)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(jsonOutput)
		return
	}

	// Write to DB
	trashedHost := Garbage{
		Host:    myIID.InstanceID,
		Region:  myIID.Region,
		Account: myIID.AccountID,
		Role:    ec2role,
		Type:    ec2type,
	}

	if err := db.Create(&trashedHost).Error; err != nil {
		// I don't love this, but the gorm "errors.Is" doesn't have something for this, and in this case
		// it's okay -- multiple people may `sudo nt` a host
		if !strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			resp := response{Accepted: false, Context: err.Error()}
			jsonOutput, _ := json.Marshal(&resp)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(jsonOutput)
			return
		}
	}
	log.Infof("/v1/trash/new: ACK %s:%s(%s/%s)%s",
		conf.Accounts[myIID.AccountID].Name, trashedHost.Host, trashedHost.Role,
		trashedHost.Type, trashedHost.Region)
	resp := response{Accepted: true}
	jsonOutput, _ := json.Marshal(&resp)
	w.Write(jsonOutput)
}

func verifyIID(doc []byte, rsaSigRaw []byte) error {
	if len(rsaSigRaw) == 0 {
		return fmt.Errorf("signature is empty")
	}
	RSASig, err := base64.StdEncoding.DecodeString(string(rsaSigRaw))
	if err != nil {
		return fmt.Errorf("unable to un-base64 the instance certificate")
	}
	iidVerifierRaw, err := ioutil.ReadFile(conf.IIDCert)
	if err != nil {
		panic(err)
	}
	RSACertPEM, _ := pem.Decode(iidVerifierRaw)
	RSACert, err := x509.ParseCertificate(RSACertPEM.Bytes)
	if err != nil {
		panic(err)
	}

	validate := RSACert.CheckSignature(x509.SHA256WithRSA, doc, RSASig)
	if validate != nil {
		return err
	}
	return nil
}
