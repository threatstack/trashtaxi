// garbaged - clean up manually modified hosts, quick
// PickupTrashHandler.go: Removes old hosts.
//
// Copyright 2018-2022 F5 Inc.
// Licensed under the BSD 3-clause license; see LICENSE.md for more information.

package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/go-multierror"
	log "github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
)

// PickupTrash - Actually run the GC
func pickupTrash(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var errs *multierror.Error
	log.Infof("/v1/trash/pickup: POST %s", r.RemoteAddr)
	funcStartTime := time.Now()

	// Only let us run this every so often to avoid a DoS
	pickupRateLimitDuration, err := time.ParseDuration(conf.TimeBetweenPickup)
	if err != nil {
		log.Warnf("Unable to convert TimeBetweenPickup to time.Duration")
		resp := response{Accepted: false, Context: "Unable to convert TimeBetweenPickup to time.Duration"}
		jsonOutput, _ := json.Marshal(&resp)
		w.Write(jsonOutput)
		return
	}
	if lastPickupTime.Add(pickupRateLimitDuration).After(funcStartTime) {
		log.Warnf("Pickup run in last %s - not running", conf.TimeBetweenPickup)
		resp := response{Accepted: false, Context: fmt.Sprintf("Pickup run in last %s", conf.TimeBetweenPickup)}
		jsonOutput, _ := json.Marshal(&resp)
		w.Write(jsonOutput)
		return
	}
	if pickupLock {
		log.Warnf("Pickup countdown already in progress... ")
		resp := response{Accepted: false, Context: "Pickup already running"}
		jsonOutput, _ := json.Marshal(&resp)
		w.Write(jsonOutput)
		return
	}

	pickupLock = true
	// Collect garbage
	var allTrash []Garbage
	types, roles := getHolidays()
	if err := db.Select("DISTINCT ON (role) *").Where("role NOT IN (?) AND type NOT IN (?)", roles, types).Limit(conf.AWS.TermLimit).Find(&allTrash).Error; err != nil {
		log.Warnf("Unable to pick up garbage: %v\n", err)
	}
	sortedTrash := sortGarbage(allTrash, true, true)

	// If there's no garbage to pick up, might as well exit gracefully now.
	if len(sortedTrash) == 0 {
		log.Warnf("No garbage to pick up, exiting.")
		resp := response{Accepted: true, Context: "Nothing to pickup"}
		jsonOutput, _ := json.Marshal(&resp)
		w.Write(jsonOutput)
		lastPickupTime = funcStartTime
		pickupLock = false

		// If slack is enabled, send a message so folks know the taxi's still running
		if conf.Slack.Enabled {
			api := slack.New(conf.Slack.APIKey)
			var msg []string
			msg = append(msg, ":wave: No trash to take out, have a nice day!")

			opts := slack.MsgOptionCompose(
				slack.MsgOptionAsUser(true),
				slack.MsgOptionDisableLinkUnfurl(),
				slack.MsgOptionText(strings.Join(msg, "\n"), false),
			)

			for _, channel := range conf.Slack.Channels {
				channelID, timestamp, err := api.PostMessage(channel, opts)
				if err != nil {
					log.Warnf("%s\n", err)
				}
				log.Infof("Message successfully sent to channel %s(%s) at %s", channel, channelID, timestamp)
			}
		}

		return
	}

	// If slack is enabled, let's warn folks about whats about to happen
	if conf.Slack.Enabled {
		api := slack.New(conf.Slack.APIKey)
		var msg []string
		msg = append(msg, ":wave: I'm going to terminate these hosts:")
		for _, v := range sortedTrash {
			msg = append(msg, fmt.Sprintf("> %s, a _%s_ host in *%s* (%s)", v.Host, v.Role, conf.Accounts[v.Account].Name, v.Region))
		}
		msgTimeZone, _ := time.LoadLocation("America/New_York")
		msgTime := time.Now().In(msgTimeZone)
		timeWaitDuration, err := time.ParseDuration(conf.Slack.TimeWait)
		if err != nil {
			log.Warnf("Unable to parse time: %s", err)
		}
		msgTime = msgTime.Add(timeWaitDuration)
		msg = append(msg, fmt.Sprintf("You have until %s (%s) to restart `garbaged` if you wish to stop this process.", msgTime.Format("15:04:05 MST"), conf.Slack.TimeWait))
		opts := slack.MsgOptionCompose(
			slack.MsgOptionAsUser(true),
			slack.MsgOptionDisableLinkUnfurl(),
			slack.MsgOptionText(strings.Join(msg, "\n"), false),
		)

		for _, channel := range conf.Slack.Channels {
			channelID, timestamp, err := api.PostMessage(channel, opts)
			if err != nil {
				log.Warnf("%s\n", err)
			}
			log.Infof("Message successfully sent to channel %s(%s) at %s", channel, channelID, timestamp)
		}

		// Also if we're using slack, we have a built-in timer that lets people
		// react to maybe some hosts not being what we want. We need to write out the
		// 200 to the TT client now, otherwise it'll be a while.
		resp := response{Accepted: true}
		jsonOutput, _ := json.Marshal(&resp)
		w.Write(jsonOutput)

		// Theres a possibility that we may want to stop this taxicab, so...
		log.Infof("Sleeping for %s before continuing...", conf.Slack.TimeWait)
		slackNotificationWait, err := time.ParseDuration(conf.Slack.TimeWait)
		if err != nil {
			log.Warnf("Unable to convert TimeBetweenPickup to time.Duration")
			return
		}
		time.Sleep(slackNotificationWait)
		log.Infof("Slack notification wait passed, continuing...")
	}

	// different regions will have different creds, so let's store them somewhere
	var awsCredentialCache map[string]*credentials.Credentials = make(map[string]*credentials.Credentials)

	for _, v := range sortedTrash {
		log.Infof("/v1/trash/pickup: Picking up %s:%s(%s/%s)%s\n",
			conf.Accounts[v.Account].Name, v.Host, v.Role, v.Type, v.Region)
		awsSession, err := awsSession(v.Region)
		if err != nil {
			text := fmt.Sprintf("Unable to start AWS session (%s)", err)
			log.Warnf("/v1/trash/pickup: %s", text)
			errs = multierror.Append(errs, fmt.Errorf("%s", text))
		}

		if awsCredentialCache[v.Account] == nil {
			awsCredentialCache[v.Account] = awsStsCreds(conf.Accounts[v.Account].ARN,
				conf.Accounts[v.Account].ExternalID, awsSession)
			if _, err := awsCredentialCache[v.Account].Get(); err != nil {
				text := fmt.Sprintf("Unable to get AWS STS Credentials (%s)", err)
				log.Warnf("/v1/trash/pickup: %s", text)
				errs = multierror.Append(errs, fmt.Errorf("%s", text))
			}
		}
		svc := ec2.New(awsSession, &aws.Config{Credentials: awsCredentialCache[v.Account]})
		termReq := &ec2.TerminateInstancesInput{
			DryRun:      aws.Bool(conf.AWS.DryRun),
			InstanceIds: aws.StringSlice([]string{v.Host})}

		_, err = svc.TerminateInstances(termReq)
		if err != nil {
			if aerr, ok := err.(awserr.Error); ok {
				// If the instance is already gone then move on (but delete it from the database)
				if aerr.Code() == "InvalidInstanceID.NotFound" {
					text := fmt.Sprintf("Instance not found, moving on! (%s)", err)
					log.Warnf("/v1/trash/pickup: %s", text)
					errs = multierror.Append(errs, fmt.Errorf("%s", text))
				} else {
					text := fmt.Sprintf("Unable to Terminate (%s)", err)
					log.Warnf("/v1/trash/pickup: %s", text)
					errs = multierror.Append(errs, fmt.Errorf("%s", text))
					continue
				}
			} else {
				text := fmt.Sprintf("Unable to Terminate (%s)", err)
				log.Warnf("/v1/trash/pickup: %s", text)
				errs = multierror.Append(errs, fmt.Errorf("%s", text))
				continue
			}
		}
		log.Infof("/v1/trash/pickup: TerminateInstance sent for %s ", v.Host)

		data := Garbage{Host: v.Host}
		if err := db.Where("host LIKE ?", v.Host).Delete(&data).Error; err != nil {
			text := fmt.Sprintf("Unable to delete host from database (%s)", err)
			log.Warnf("/v1/trash/pickup: %s", text)
			errs = multierror.Append(errs, fmt.Errorf("%s", text))
		}
		log.Debugf("/v1/trash/pickup: %s removed from database", v.Host)
	}

	// If there's no slack, handle errors like we normally would.
	if !conf.Slack.Enabled {
		resp := response{Accepted: true}
		if errs != nil {
			resp.Context = errs.Error()
		}
		jsonOutput, _ := json.Marshal(&resp)
		w.Write(jsonOutput)
	}

	// We made it everyone. Lets make sure we cant be triggered unless
	// we're > the minimum time for that.
	if errs == nil {
		lastPickupTime = funcStartTime
		pickupLock = false
	}
}
