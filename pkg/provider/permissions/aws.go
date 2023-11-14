package permissions

import (
	"context"
	"fmt"
	"regexp"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	plrlErrors "github.com/pluralsh/plural-cli/pkg/utils/errors"
)

type AwsChecker struct {
	ctx context.Context
	cfg aws.Config
}

var (
	awsExpected = []string{
		"eks:CreateCluster",
		"eks:CreateNodeGroup",
		"eks:CreateAddOn",
		"s3:CreateBucket",
		"vpc:CreateVpc",
		"iam:CreateRole",
		"iam:CreateOpenIDConnectProvider",
	}
	roleRegex = regexp.MustCompile(`assumed-role/([\w+=,.@-]+)/`)
	rootRegex = regexp.MustCompile(`arn:aws:iam::[0-9A-Z]+:root`)
)

func NewAwsChecker(ctx context.Context) (*AwsChecker, error) {
	cfg, err := awsConfig.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, plrlErrors.ErrorWrap(err, "could not instantiate aws client: ")
	}
	return &AwsChecker{ctx, cfg}, nil
}

func (c *AwsChecker) getOriginalIdentity(arn string) (string, error) {
	match := roleRegex.FindStringSubmatch(arn)
	if match == nil {
		return arn, nil
	}

	iamSvc := iam.NewFromConfig(c.cfg)
	role, err := iamSvc.GetRole(c.ctx, &iam.GetRoleInput{RoleName: aws.String(match[1])})
	if err != nil {
		return "", err
	}

	return *role.Role.Arn, nil
}

func (c *AwsChecker) MissingPermissions() (result []string, err error) {
	svc := sts.NewFromConfig(c.cfg)
	id, err := svc.GetCallerIdentity(c.ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return
	}

	iamSvc := iam.NewFromConfig(c.cfg)
	arn, err := c.getOriginalIdentity(*id.Arn)
	if err != nil {
		return
	}

	if rootRegex.MatchString(arn) {
		err = fmt.Errorf("It looks like your aws identity %s is the root user of your account. Using the root user is highly insecure, we recommend creating an iam user and using that instead", arn)
		return
	}

	resp, err := iamSvc.SimulatePrincipalPolicy(c.ctx, &iam.SimulatePrincipalPolicyInput{
		PolicySourceArn: aws.String(arn),
		ActionNames:     awsExpected,
	})
	if err != nil {
		return
	}

	result = make([]string, 0)
	for _, res := range resp.EvaluationResults {
		if res.EvalDecision != types.PolicyEvaluationDecisionTypeAllowed {
			result = append(result, *res.EvalActionName)
		}
	}

	return
}
