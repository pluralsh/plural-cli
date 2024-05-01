package plural

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/pluralsh/plural-cli/pkg/kubernetes"
	"github.com/pluralsh/plural-cli/pkg/manifest"
	"github.com/pluralsh/plural-cli/pkg/provider"
	"github.com/pluralsh/plural-cli/pkg/utils"
	"github.com/urfave/cli"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	clusterapioperator "sigs.k8s.io/cluster-api-operator/api/v1alpha1"
	clusterapi "sigs.k8s.io/cluster-api/api/v1beta1"
	apiclient "sigs.k8s.io/cluster-api/cmd/clusterctl/client"
	ctrlruntimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

var runtimescheme = runtime.NewScheme()

func init() {
	utilruntime.Must(corev1.AddToScheme(runtimescheme))
	utilruntime.Must(apiextensionsv1.AddToScheme(runtimescheme))
	utilruntime.Must(clusterapi.AddToScheme(runtimescheme))
	utilruntime.Must(clusterapioperator.AddToScheme(runtimescheme))
}

const (
	kindConfig = `kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
networking:
  ipFamily: dual
nodes:
- role: control-plane
  extraMounts:
    - hostPath: /var/run/docker.sock
      containerPath: /var/run/docker.sock`
)

func (p *Plural) bootstrapCommands() []cli.Command {
	return []cli.Command{
		{
			Name:        "cluster",
			Subcommands: p.bootstrapClusterCommands(),
			Usage:       "Manage bootstrap cluster",
		},
		{
			Name:        "namespace",
			Subcommands: p.namespaceCommands(),
			Usage:       "Manage bootstrap cluster",
		},
	}
}

func (p *Plural) namespaceCommands() []cli.Command {
	return []cli.Command{
		{
			Name:      "create",
			ArgsUsage: "NAME",
			Usage:     "Creates bootstrap namespace",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "skip-if-exists",
					Usage: "skip creating when namespace exists",
				},
			},
			Action: latestVersion(initKubeconfig(requireArgs(p.handleCreateNamespace, []string{"NAME"}))),
		},
	}
}

func (p *Plural) bootstrapClusterCommands() []cli.Command {
	return []cli.Command{
		{
			Name:      "create",
			ArgsUsage: "NAME",
			Usage:     "Creates bootstrap cluster",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "image",
					Usage: "kind image to use",
				},
				cli.BoolFlag{
					Name:  "skip-if-exists",
					Usage: "skip creating when cluster exists",
				},
			},
			Action: latestVersion(requireKind(requireArgs(handleCreateCluster, []string{"NAME"}))),
		},
		{
			Name:      "delete",
			ArgsUsage: "NAME",
			Usage:     "Deletes bootstrap cluster",
			Action:    latestVersion(requireKind(requireArgs(handleDeleteCluster, []string{"NAME"}))),
		},
		{
			Name:  "move",
			Usage: "Move cluster API objects",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "kubeconfig",
					Usage: "path to the kubeconfig file for the source management cluster. If unspecified, default discovery rules apply.",
				},
				cli.StringFlag{
					Name:  "kubeconfig-context",
					Usage: "context to be used within the kubeconfig file for the source management cluster. If empty, current context will be used.",
				},
				cli.StringFlag{
					Name:  "to-kubeconfig",
					Usage: "path to the kubeconfig file to use for the destination management cluster.",
				},
				cli.StringFlag{
					Name:  "to-kubeconfig-context",
					Usage: "Context to be used within the kubeconfig file for the destination management cluster. If empty, current context will be used.",
				},
			},
			Action: latestVersion(p.handleMoveCluster),
		},
		{
			Name:      "destroy-cluster-api",
			ArgsUsage: "NAME",
			Usage:     "Destroy cluster API",
			Action:    latestVersion(requireArgs(p.handleDestroyClusterAPI, []string{"NAME"})),
		},
	}
}

func (p *Plural) handleDestroyClusterAPI(c *cli.Context) error {
	name := c.Args().Get(0)
	_, found := utils.ProjectRoot()
	if !found {
		return fmt.Errorf("You're not within an installation repo")
	}
	pm, err := manifest.FetchProject()
	if err != nil {
		return err
	}
	prov := &provider.KINDProvider{Clust: "bootstrap"}
	if err := prov.KubeConfig(); err != nil {
		return err
	}
	config, err := kubernetes.KubeConfig()
	if err != nil {
		return err
	}
	client, err := genClientFromConfig(config)
	if err != nil {
		return err
	}
	utils.Warn("Waiting for the operator ")
	if err := utils.WaitFor(20*time.Minute, 10*time.Second, func() (bool, error) {
		pods := &corev1.PodList{}
		providerName := pm.Provider
		if providerName == "kind" {
			providerName = "docker"
		}
		selector := fmt.Sprintf("infrastructure-%s", strings.ToLower(providerName))
		if err := client.List(context.Background(), pods, ctrlruntimeclient.MatchingLabels{"cluster.x-k8s.io/provider": selector}); err != nil {
			if !apierrors.IsNotFound(err) {
				return false, fmt.Errorf("failed to get pods: %w", err)
			}
			return false, nil
		}
		if len(pods.Items) > 0 {
			if isReady(pods.Items[0].Status.Conditions) {
				return true, nil
			}
		}
		utils.Warn(".")
		return false, nil
	}); err != nil {
		return err
	}
	if err := client.Delete(context.Background(), &clusterapi.Cluster{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: "bootstrap"},
	}); err != nil {
		return err
	}
	utils.Warn("\nDeleting cluster")
	return utils.WaitFor(40*time.Minute, 10*time.Second, func() (bool, error) {
		if err := client.Get(context.Background(), ctrlruntimeclient.ObjectKey{Name: name, Namespace: "bootstrap"}, &clusterapi.Cluster{}); err != nil {
			if !apierrors.IsNotFound(err) {
				return false, fmt.Errorf("failed to get Cluster: %w", err)
			}
			return true, nil
		}
		utils.Warn(".")
		return false, nil
	})
}

func isReady(conditions []corev1.PodCondition) bool {
	for _, cond := range conditions {
		if cond.Type == corev1.PodReady && cond.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}

func (p *Plural) handleMoveCluster(c *cli.Context) error {
	_, found := utils.ProjectRoot()
	if !found {
		return fmt.Errorf("You're not within an installation repo")
	}

	client, err := apiclient.New("")
	if err != nil {
		return err
	}

	kubeconfig := c.String("kubeconfig")
	kubeconfigContext := c.String("kubeconfig-context")
	toKubeconfig := c.String("to-kubeconfig")
	toKubeconfigContext := c.String("to-kubeconfig-context")

	options := apiclient.MoveOptions{
		FromKubeconfig: apiclient.Kubeconfig{
			Path:    kubeconfig,
			Context: kubeconfigContext,
		},
		ToKubeconfig: apiclient.Kubeconfig{
			Path:    toKubeconfig,
			Context: toKubeconfigContext,
		},
		Namespace: "bootstrap",
		DryRun:    false,
	}
	if err := client.Move(options); err != nil {
		return err
	}

	return nil
}

func (p *Plural) handleCreateNamespace(c *cli.Context) error {
	name := c.Args().Get(0)
	fmt.Printf("Creating namespace %s ...\n", name)
	err := p.InitKube()
	if err != nil {
		return err
	}
	if err := p.CreateNamespace(name, true); err != nil {
		if apierrors.IsAlreadyExists(err) {
			return nil
		}
		return err
	}

	return nil
}

func handleDeleteCluster(c *cli.Context) error {
	name := c.Args().Get(0)
	return utils.Exec("kind", "delete", "cluster", "--name", name)
}

func handleCreateCluster(c *cli.Context) error {
	name := c.Args().Get(0)
	imageFlag := c.String("image")
	skipCreation := c.Bool("skip-if-exists")
	if utils.IsKindClusterAlreadyExists(name) && skipCreation {
		utils.Highlight("Cluster %s already exists \n", name)
		return nil
	}

	dir, err := os.MkdirTemp("", "kind")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)
	config := path.Join(dir, "config.yaml")
	if err := os.WriteFile(config, []byte(kindConfig), 0644); err != nil {
		return err
	}
	args := []string{"create", "cluster", "--name", name, "--config", config}
	if imageFlag != "" {
		args = append(args, "--image", imageFlag)
	}

	if err := utils.Exec("kind", args...); err != nil {
		return err
	}

	kubeconfig, err := utils.GetKindClusterKubeconfig(name, false)
	if err != nil {
		return err
	}

	client, err := getClient(kubeconfig)
	if err != nil {
		return err
	}
	if err := client.Create(context.Background(), &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "bootstrap",
		},
	}); err != nil {
		return err
	}
	internalKubeconfig, err := utils.GetKindClusterKubeconfig(name, true)
	if err != nil {
		return err
	}
	kubeconfigSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kubeconfig",
			Namespace: "bootstrap",
		},
		Data: map[string][]byte{
			"value": []byte(internalKubeconfig),
		},
	}
	if err := client.Create(context.Background(), kubeconfigSecret); err != nil {
		return err
	}

	return nil
}

func getClient(rawKubeconfig string) (ctrlruntimeclient.Client, error) {

	cfg, err := clientcmd.Load([]byte(rawKubeconfig))
	if err != nil {
		return nil, err
	}
	clientConfig, err := getRestConfig(cfg)
	if err != nil {
		return nil, err
	}

	return genClientFromConfig(clientConfig)
}

func genClientFromConfig(cfg *rest.Config) (ctrlruntimeclient.Client, error) {
	return ctrlruntimeclient.New(cfg, ctrlruntimeclient.Options{
		Scheme: runtimescheme,
	})
}

func getRestConfig(cfg *clientcmdapi.Config) (*rest.Config, error) {
	iconfig := clientcmd.NewNonInteractiveClientConfig(
		*cfg,
		"",
		&clientcmd.ConfigOverrides{},
		nil,
	)

	clientConfig, err := iconfig.ClientConfig()
	if err != nil {
		return nil, err
	}

	// Avoid blocking of the controller by increasing the QPS for user cluster interaction
	clientConfig.QPS = 20
	clientConfig.Burst = 50

	return clientConfig, nil
}
