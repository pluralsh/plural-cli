package gcp

import (
	"context"
	"fmt"
	"strings"

	serviceusage "cloud.google.com/go/serviceusage/apiv1"
	"cloud.google.com/go/serviceusage/apiv1/serviceusagepb"
)

type BucketLocation string

const (
	BucketLocationUS   BucketLocation = "US"
	BucketLocationEU   BucketLocation = "EU"
	BucketLocationASIA BucketLocation = "ASIA"
)

func getBucketLocation(region string) BucketLocation {
	reg := strings.ToLower(region)

	if strings.Contains(reg, "us") ||
		strings.Contains(reg, "northamerica") ||
		strings.Contains(reg, "southamerica") {
		return BucketLocationUS
	}

	if strings.Contains(reg, "europe") {
		return BucketLocationEU
	}

	if strings.Contains(reg, "asia") ||
		strings.Contains(reg, "australia") {
		return BucketLocationASIA
	}

	return BucketLocationUS
}

func tryToEnableServices(ctx context.Context, client *serviceusage.Client, req *serviceusagepb.BatchEnableServicesRequest) (err error) {
	op, err := client.BatchEnableServices(ctx, req)
	if err != nil {
		return
	}

	_, err = op.Wait(ctx)
	return
}

func printUserInfo() error {
	email, name, err := LoggedInUserInfo()
	if err != nil {
		return err
	}

	fmt.Print("\nUsing GCP Credentials\n")
	fmt.Printf("User email: %s\n", email)

	if len(name) > 0 {
		fmt.Printf("User name: %s\n", name)
	}

	fmt.Println()
	return nil
}
