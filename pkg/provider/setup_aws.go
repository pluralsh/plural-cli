package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/samber/lo"
)

type awsSetup struct{}

func (awsSetup) Name() string { return "aws" }

func (awsSetup) Schema() []SetupField {
	return []SetupField{
		{Key: "cluster", Label: "Cluster name", Placeholder: "max 15 chars", Required: true},
		{Key: "region", Label: "Region", Placeholder: "e.g. us-east-2", Default: "us-east-2", Required: true},
	}
}

func (s awsSetup) Probe(ctx context.Context) (SetupResult, error) {
	_, identity, err := GetAWSCallerIdentity(ctx)
	if err != nil {
		return SetupResult{}, fmt.Errorf("AWS credentials: %w", err)
	}
	return SetupResult{
		Summary: fmt.Sprintf("AWS profile %s · account %s · %s",
			AWSProfileName(),
			lo.FromPtr(identity.Account),
			truncateMiddle(lo.FromPtr(identity.Arn), 48),
		),
		Fields: withOptions(s.Schema(), "region", AWSRegions()),
	}, nil
}

func (awsSetup) Options(_ context.Context, fieldKey string, _ map[string]string) ([]string, error) {
	if fieldKey == "region" {
		return AWSRegions(), nil
	}
	return nil, nil
}

// Preflights runs the same IAM permission check as AWSProvider.Preflights().
func (awsSetup) Preflights(ctx context.Context, values map[string]string) error {
	iamSession, _, err := GetAWSCallerIdentity(ctx)
	if err != nil {
		return err
	}
	prov := &AWSProvider{
		Clus:      strings.TrimSpace(values["cluster"]),
		Reg:       strings.TrimSpace(values["region"]),
		goContext: &ctx,
		ctx:       map[string]any{"IAMSession": iamSession},
	}
	for _, pre := range prov.Preflights() {
		if err := pre.Validate(); err != nil {
			return err
		}
	}
	return nil
}
