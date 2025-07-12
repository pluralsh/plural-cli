// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"strings"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
)

// RoleNameSessionFromARN returns the role and session names in an ARN if any.
// Otherwise, it returns empty strings.
func RoleNameSessionFromARN(rawARN string) (string, string) {
	parsedARN, err := arn.Parse(rawARN)

	if err != nil {
		return "", ""
	}

	reAssume := regexache.MustCompile(`^assumed-role/.{1,}/.{2,}`)

	if !reAssume.MatchString(parsedARN.Resource) || parsedARN.Service != "sts" {
		return "", ""
	}

	parts := strings.Split(parsedARN.Resource, "/")

	if len(parts) < 3 {
		return "", ""
	}

	return parts[len(parts)-2], parts[len(parts)-1]
}
