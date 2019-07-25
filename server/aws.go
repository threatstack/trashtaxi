// garbaged - clean up manually modified hosts, quick
// aws.go: stuff to make AWS work
//
// Copyright 2018 Threat Stack, Inc.
// Licensed under the BSD 3-clause license; see LICENSE.md for more information.
// Author: Patrick T. Cable II <pat.cable@threatstack.com>

package server

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/go-multierror"
)

func awsSession(region string) (sess *session.Session, err error) {
	sess, err = session.NewSession(&aws.Config{
		Region: aws.String(region)},
	)
	return
}

func awsStsCreds(arn string, externalID string, sess *session.Session) (creds *credentials.Credentials) {
	creds = stscreds.NewCredentials(sess, arn,
		func(p *stscreds.AssumeRoleProvider) {
			p.ExternalID = aws.String(externalID)
		})
	return
}

func getTags(svc *ec2.EC2, instance string, roletag string, typetag string) (string, string, error) {
	var funcErr *multierror.Error
	var ec2role string
	var ec2type string
	var foundrole bool
	var foundtype bool

	tagSearch := &ec2.DescribeTagsInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("resource-id"),
				Values: []*string{
					aws.String(instance),
				},
			},
		},
	}
	tags, err := svc.DescribeTags(tagSearch)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				multierror.Append(funcErr, fmt.Errorf("%s", aerr.Error()))
				return "", "", funcErr.ErrorOrNil()
			}
		} else {
			multierror.Append(funcErr, fmt.Errorf("%s", err.Error()))
			return "", "", funcErr.ErrorOrNil()
		}
	}

	for _, obj := range tags.Tags {
		if aws.StringValue(obj.Key) == roletag {
			ec2role = aws.StringValue(obj.Value)
			foundrole = true
		}
		if aws.StringValue(obj.Key) == typetag {
			ec2type = aws.StringValue(obj.Value)
			foundtype = true
		}
	}

	if foundrole == false {
		funcErr = multierror.Append(funcErr,
			fmt.Errorf("Unable to find a Role(tag '%s') for %s", roletag, instance))
	}
	if foundtype == false {
		funcErr = multierror.Append(funcErr,
			fmt.Errorf("Unable to find a Type(tag '%s') for %s", typetag, instance))
	}

	return ec2role, ec2type, funcErr.ErrorOrNil()
}
