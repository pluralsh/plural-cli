package gcp

import (
	"context"
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
	} else if strings.Contains(reg, "europe") {
		return BucketLocationEU
	} else if strings.Contains(reg, "asia") ||
		strings.Contains(reg, "australia") {
		return BucketLocationASIA
	} else {
		return BucketLocationUS
	}
}

func tryToEnableServices(ctx context.Context, client *serviceusage.Client, req *serviceusagepb.BatchEnableServicesRequest) (err error) {
	op, err := client.BatchEnableServices(ctx, req)
	if err != nil {
		return
	}

	_, err = op.Wait(ctx)
	return
}
