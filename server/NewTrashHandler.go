// garbaged - clean up manually modified hosts, quick
// NewTrashHandler.go: The handler for trash/new
//
// Copyright 2018 Threat Stack, Inc.
// Licensed under the BSD 3-clause license; see LICENSE.md for more information.
// Author: Patrick T. Cable II <pat.cable@threatstack.com>

package server

import (
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/service/ec2"
	log "github.com/sirupsen/logrus"
	"go.mozilla.org/pkcs7"
)

type incomingTrash struct {
	IID string `json:"iid"`
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

	// Perform verification. VerifyPKCS7Doc is a scary function.
	myIID, err := VerifyPKCS7Doc(trash.IID, conf.IIDCert)
	if err != nil {
		log.Warnf("/v1/trash/new: Unable to VerifyPKCS7Doc")
		resp := response{Accepted: false, Context: "Unable to verify PKCS7 Document"}
		jsonOutput, _ := json.Marshal(&resp)
		w.Write(jsonOutput)
		return
	}

	// Do we know about this account? If not, let's not accept garbage from it.
	if stringInSlice(myIID.AccountID, knownAWSAccts) == false {
		text := fmt.Sprintf("Cant collect hosts in unknown account (account %s not configured)", myIID.AccountID)
		log.Warnf("/v1/trash/new: %s", text)
		resp := response{Accepted: false, Context: text}
		jsonOutput, _ := json.Marshal(&resp)
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
		resp := response{Accepted: false, Context: err.Error()}
		jsonOutput, _ := json.Marshal(&resp)
		w.Write(jsonOutput)
		return
	}
	log.Infof("/v1/trash/new: ACK %s:%s(%s/%s)%s",
		conf.Accounts[myIID.AccountID].Name, trashedHost.Host, trashedHost.Role,
		trashedHost.Type, trashedHost.Region)
	resp := response{Accepted: true}
	jsonOutput, _ := json.Marshal(&resp)
	w.Write(jsonOutput)
	return
}

// VerifyPKCS7Doc - Ugliness that will take an IID and validator certificate
// and return accordingly.
func VerifyPKCS7Doc(rawIID string, validator string) (parsedIID ec2metadata.EC2InstanceIdentityDocument, err error) {
	// Set up verifier certificate - read file, make it a PEM object, then a
	// *x509.Certificate. Append to an array for use later.
	iidVerifierRaw, err := ioutil.ReadFile(validator)
	if err != nil {
		panic(err)
	}
	var iidVerifiers []*x509.Certificate
	iidVerifierASN, _ := pem.Decode(iidVerifierRaw)
	iidVerifierCrt, err := x509.ParseCertificate(iidVerifierASN.Bytes)
	if err != nil {
		panic(err)
	}
	iidVerifiers = append(iidVerifiers, iidVerifierCrt)

	// Pull in the information from the client - lot of parsing and whanot here
	pkcs7B64 := fmt.Sprintf("-----BEGIN PKCS7-----\n%s\n-----END PKCS7-----",
		rawIID)
	pkcs7BER, pkcs7Rest := pem.Decode([]byte(pkcs7B64))
	if len(pkcs7Rest) != 0 {
		err = fmt.Errorf("pem.Decode failed: %s", err)
		return
	}
	pkcs7Data, err := pkcs7.Parse(pkcs7BER.Bytes)
	if err != nil {
		err = fmt.Errorf("pkcs7.Parse failed: %s", err)
		return
	}
	// Configure verifiers
	pkcs7Data.Certificates = iidVerifiers

	if pkcs7Data.Verify() != nil {
		err = fmt.Errorf("pkcs7.Verify failed :(")
	} else {
		json.Unmarshal(pkcs7Data.Content, &parsedIID)
	}
	return
}
