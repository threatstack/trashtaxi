// garbaged - clean up manually modified hosts, quick
// CleanupTrashHandler.go: The handler for all GET trash/ endpoints
//
// Copyright 2018 Threat Stack, Inc.
// Licensed under the BSD 3-clause license; see LICENSE.md for more information.
// Author: Patrick T. Cable II <pat.cable@threatstack.com>

package server

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/ec2"
	multierror "github.com/hashicorp/go-multierror"
	log "github.com/sirupsen/logrus"
)

type cleanupResponse struct {
	Accepted bool                           `json:"accepted"`
	Nodes    map[string]map[string][]string `json:"nodes"`
	Context  string                         `json:"context"`
}

func cleanupTrash(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var errs *multierror.Error
	var allTrash []Garbage
	var awsCredentialCache map[string]*credentials.Credentials
	awsCredentialCache = make(map[string]*credentials.Credentials)

	if err := db.Find(&allTrash).Error; err != nil {
		fmt.Printf("no: %v\n", err)
	}

	// Sort instances into a data stucture that lets us collect instances per account
	var instances map[string]map[string][]string
	instances = make(map[string]map[string][]string)
	var allRemoved map[string]map[string][]string
	allRemoved = make(map[string]map[string][]string)
	for _, v := range allTrash {
		if instances[v.Region] == nil {
			instances[v.Region] = map[string][]string{}
			allRemoved[v.Region] = map[string][]string{}
		}
		if instances[v.Region][v.Account] == nil {
			instances[v.Region][v.Account] = []string{}
			allRemoved[v.Region][v.Account] = []string{}
		}
		instances[v.Region][v.Account] = append(instances[v.Region][v.Account], v.Host)
	}

	// ... so that we can group them into one big happy DescribeInstance call. By region. And account ID.
	search := regexp.MustCompile("i-[0-9a-z]{17}")
	for region, instancegroup := range instances {
		for account, instances := range instancegroup {
			// set up session for this action
			awsSession, err := awsSession(region)
			if err != nil {
				text := fmt.Sprintf("Unable to start AWS session (%s)", err)
				log.Warnf("/v1/trash/cleanup: %s", text)
				errs = multierror.Append(errs, fmt.Errorf("%s", text))
			}
			if awsCredentialCache[account] == nil {
				awsCredentialCache[account] = awsStsCreds(conf.Accounts[account].ARN,
					conf.Accounts[account].ExternalID, awsSession)
				if _, err := awsCredentialCache[account].Get(); err != nil {
					text := fmt.Sprintf("Unable to get AWS STS Credentials (%s)", err)
					log.Warnf("/v1/trash/cleanup: %s", text)
					errs = multierror.Append(errs, fmt.Errorf("%s", text))
					continue
				}
			}
			var deadInstances []string

			svc := ec2.New(awsSession, &aws.Config{Credentials: awsCredentialCache[account]})

			// We can only query 100 at at time.
			loops := math.Ceil(float64(len(instances)) / float64(100))
			instmin := 0
			instmax := 0
			if len(instances) >= 100 {
				instmax = 99
			} else if len(instances) > 1 {
				instmax = len(instances) - 1
			} else if len(instances) == 1 {
				// bunker only had one node, heh
				instmax = 1
			}
			for i := 1; i <= int(loops); i++ {
				reqInput := &ec2.DescribeInstanceStatusInput{
					InstanceIds: aws.StringSlice(instances[instmin:instmax]),
				}

				_, err = svc.DescribeInstanceStatus(reqInput)

				if err != nil {
					if aerr, ok := err.(awserr.Error); ok {
						if aerr.Code() == "InvalidInstanceID.NotFound" {
							// these instances dont exist. we get to parse strings, here
							// which is a very fun thing for us.
							deadInstances = append(deadInstances, search.FindAllString(aerr.Message(), -1)...)
						} else {
							text := fmt.Sprintf("Unable to Describe Instance Status (%s)", err)
							log.Warnf("/v1/trash/cleanup: %s", text)
							errs = multierror.Append(errs, fmt.Errorf("%s", text))
						}
					}
				}
				if i == (int(loops) - 1) {
					// we dont want to request more than allocated in the slice on our next run
					instmin = instmax + 1
					instmax = len(instances) - 1
				} else if i < (int(loops) - 1) {
					// we want to get another 100
					instmin = instmax + 1
					instmax = instmin + 99
				} else {
					// we're done and reset stuff
					instmin = 0
					instmax = 0
				}
			}

			// Kill old nodes from database
			for _, v := range deadInstances {
				data := Garbage{Host: v}
				if err := db.Where("host LIKE ?", v).Delete(&data).Error; err != nil {
					text := fmt.Sprintf("Unable to delete host from database (%s)", err)
					log.Warnf("/v1/trash/cleanup: %s", text)
					errs = multierror.Append(errs, fmt.Errorf("%s", text))
				}
				log.Debugf("/v1/trash/clenup: %s removed from database", v)
				allRemoved[region][account] = append(allRemoved[region][account], v)
			}
		}
	}

	resp := cleanupResponse{Accepted: true, Nodes: allRemoved}
	if errs != nil {
		resp.Context = errs.Error()
	}
	jsonOutput, _ := json.Marshal(&resp)
	w.Write(jsonOutput)
	return
}

func remove(s []string, r string) []string {
	for i, v := range s {
		if v == r {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}
