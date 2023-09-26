package aws

import (
	"context"
	"fmt"
	"strings"

	"github.com/pluralsh/plural/pkg/kubernetes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ekscontrolplanev1 "sigs.k8s.io/cluster-api-provider-aws/v2/controlplane/eks/api/v1beta2"
	"sigs.k8s.io/yaml"
)

const (
	awsAuthNs   = "kube-system"
	awsAuthName = "aws-auth"
	roleKey     = "mapRoles"
	usersKey    = "mapUsers"
)

func FetchAuth() (*ekscontrolplanev1.IAMAuthenticatorConfig, error) {
	ctx := context.Background()
	kube, err := kubernetes.Kubernetes()
	if err != nil {
		return nil, err
	}

	return fetchAwsAuth(ctx, kube)
}

func AddUser(userArn string) error {
	ctx := context.Background()
	kube, err := kubernetes.Kubernetes()
	if err != nil {
		return err
	}

	eksConfig, err := fetchAwsAuth(ctx, kube)
	if err != nil {
		return err
	}

	eksConfig.UserMappings = append(eksConfig.UserMappings, ekscontrolplanev1.UserMapping{
		KubernetesMapping: ekscontrolplanev1.KubernetesMapping{
			Groups:   []string{"system:masters"},
			UserName: username(userArn),
		},
		UserARN: userArn,
	})

	return persistAuth(ctx, kube, eksConfig)
}

func AddRole(roleArn string) error {
	ctx := context.Background()
	kube, err := kubernetes.Kubernetes()
	if err != nil {
		return err
	}

	eksConfig, err := fetchAwsAuth(ctx, kube)
	if err != nil {
		return err
	}

	eksConfig.RoleMappings = append(eksConfig.RoleMappings, ekscontrolplanev1.RoleMapping{
		KubernetesMapping: ekscontrolplanev1.KubernetesMapping{
			Groups:   []string{"system:masters"},
			UserName: username(roleArn),
		},
		RoleARN: roleArn,
	})

	return persistAuth(ctx, kube, eksConfig)
}

func username(arn string) string {
	parts := strings.Split(arn, "/")
	return parts[len(parts)-1]
}

func fetchAwsAuth(ctx context.Context, kube kubernetes.Kube) (*ekscontrolplanev1.IAMAuthenticatorConfig, error) {
	client := kube.GetClient()
	cm, err := client.CoreV1().ConfigMaps(awsAuthNs).Get(ctx, awsAuthName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	res := &ekscontrolplanev1.IAMAuthenticatorConfig{
		RoleMappings: []ekscontrolplanev1.RoleMapping{},
		UserMappings: []ekscontrolplanev1.UserMapping{},
	}

	if rolesSection, ok := cm.Data[roleKey]; ok {
		err := yaml.Unmarshal([]byte(rolesSection), &res.RoleMappings)
		if err != nil {
			return nil, fmt.Errorf("unmarshalling mapped roles: %w", err)
		}
	}

	if usersSection, ok := cm.Data[usersKey]; ok {
		err := yaml.Unmarshal([]byte(usersSection), &res.UserMappings)
		if err != nil {
			return nil, fmt.Errorf("unmarshalling mapped users: %w", err)
		}
	}

	return res, nil
}

func persistAuth(ctx context.Context, kube kubernetes.Kube, authConfig *ekscontrolplanev1.IAMAuthenticatorConfig) error {
	client := kube.GetClient()
	cmClient := client.CoreV1().ConfigMaps(awsAuthNs)
	cm, err := cmClient.Get(ctx, awsAuthName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	if len(authConfig.RoleMappings) > 0 {
		roleMappings, err := yaml.Marshal(authConfig.RoleMappings)
		if err != nil {
			return fmt.Errorf("marshalling auth config roles: %w", err)
		}
		cm.Data[roleKey] = string(roleMappings)
	}

	if len(authConfig.UserMappings) > 0 {
		userMappings, err := yaml.Marshal(authConfig.UserMappings)
		if err != nil {
			return fmt.Errorf("marshalling auth config users: %w", err)
		}
		cm.Data[usersKey] = string(userMappings)
	}

	_, err = cmClient.Update(ctx, cm, metav1.UpdateOptions{})
	return err
}
