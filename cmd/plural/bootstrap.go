package plural

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pluralsh/plural/pkg/api"
	"github.com/pluralsh/plural/pkg/manifest"

	"github.com/pkg/errors"
	bv1alpha1 "github.com/pluralsh/bootstrap-operator/apis/bootstrap/v1alpha1"
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
	ctrlruntime "sigs.k8s.io/controller-runtime/pkg/client"
	ctrlruntimeclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/kind/pkg/cluster"

	"github.com/pluralsh/plural/pkg/kubernetes"
	"github.com/pluralsh/plural/pkg/provider"
	"github.com/pluralsh/plural/pkg/utils"
	"github.com/pluralsh/plural/pkg/utils/pathing"
)

var runtimescheme = runtime.NewScheme()

func init() {
	utilruntime.Must(corev1.AddToScheme(runtimescheme))
	utilruntime.Must(bv1alpha1.AddToScheme(runtimescheme))
	utilruntime.Must(apiextensionsv1.AddToScheme(runtimescheme))
	utilruntime.Must(clusterapi.AddToScheme(runtimescheme))
	utilruntime.Must(clusterapioperator.AddToScheme(runtimescheme))
}

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
			Action: latestVersion(requireArgs(handleCreateCluster, []string{"NAME"})),
		},
		{
			Name:      "delete",
			ArgsUsage: "NAME",
			Usage:     "Deletes bootstrap cluster",
			Action:    latestVersion(requireArgs(handleDeleteCluster, []string{"NAME"})),
		},
		{
			Name:      "watch",
			ArgsUsage: "NAME",
			Usage:     "Watches cluster creation progress",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "enable-cluster-creation",
					Usage: "enable cluster creation",
				},
			},
			Action: latestVersion(initKubeconfig(requireArgs(p.handleWatchCluster, []string{"NAME"}))),
		},
		{
			Name:   "move",
			Usage:  "Move cluster API object to bootstrap cluster",
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
	if err := WaitFor(20*time.Minute, 10*time.Second, func() (bool, error) {
		pods := &corev1.PodList{}
		selector := fmt.Sprintf("infrastructure-%s", strings.ToLower(api.NormalizeProvider(pm.Provider)))
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
	fmt.Println()
	if err := client.Delete(context.Background(), &clusterapi.Cluster{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: "bootstrap"},
	}); err != nil {
		return err
	}
	utils.Warn("Deleting cluster ")
	return WaitFor(20*time.Minute, 10*time.Second, func() (bool, error) {
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
	prov, err := provider.GetProvider()
	if err != nil {
		return err
	}
	if err := prov.KubeConfig(); err != nil {
		return err
	}
	config, err := kubernetes.KubeConfig()
	if err != nil {
		return err
	}
	clientFrom, err := genClientFromConfig(config)
	if err != nil {
		return err
	}

	crdList := &apiextensionsv1.CustomResourceDefinitionList{}
	if err := getCRDList(context.Background(), clientFrom, crdList); err != nil {
		return err
	}
	homedir, _ := os.UserHomeDir()
	src := pathing.SanitizeFilepath(filepath.Join(homedir, ".kube", "config"))
	fileFrom, err := os.CreateTemp("", "from.config")
	if err != nil {
		return err
	}
	if err := utils.CopyFile(src, fileFrom.Name()); err != nil {
		return err
	}
	prov = &provider.KINDProvider{Clust: "bootstrap"}
	if err := prov.KubeConfig(); err != nil {
		return err
	}
	fileTo, err := os.CreateTemp("", "to.config")
	if err != nil {
		return err
	}
	if err := utils.CopyFile(src, fileTo.Name()); err != nil {
		return err
	}

	config, err = kubernetes.KubeConfig()
	if err != nil {
		return err
	}
	clientTo, err := genClientFromConfig(config)
	if err != nil {
		return err
	}
	for _, crd := range crdList.Items {
		if crd.Spec.Group != "operator.cluster.x-k8s.io" {
			copy := crd.DeepCopy()
			copy.ObjectMeta.ResourceVersion = ""
			if err := clientTo.Get(context.Background(), ctrlruntimeclient.ObjectKey{Name: copy.Name}, copy); err != nil {
				if !apierrors.IsNotFound(err) {
					return fmt.Errorf("failed to get Object(%T): %w", copy, err)
				}
				if err := clientTo.Create(context.Background(), copy); err != nil {
					return err
				}
			}
		}
	}

	client, err := apiclient.New("")
	if err != nil {
		return err
	}

	options := apiclient.MoveOptions{
		FromKubeconfig: apiclient.Kubeconfig{
			Path: fileFrom.Name(),
		},
		ToKubeconfig: apiclient.Kubeconfig{
			Path:    fileTo.Name(),
			Context: "kind-bootstrap",
		},
		Namespace: "bootstrap",
		DryRun:    false,
	}
	if err := client.Move(options); err != nil {
		return err
	}

	return nil
}

func (p *Plural) handleWatchCluster(c *cli.Context) error {
	name := c.Args().Get(0)
	enableCreation := c.Bool("enable-cluster-creation")
	if err := p.InitKube(); err != nil {
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
	var bootstrapCluster bv1alpha1.Bootstrap
	errorCount := 0
	providerReady := false
	capiOperatorReady := false
	capiOperatorComponentsReady := false
	capiCluster := false
	moveReady := false
	return WaitFor(30*time.Minute, 5*time.Second, func() (bool, error) {

		if err := client.Get(context.Background(), ctrlruntimeclient.ObjectKey{Name: name, Namespace: "bootstrap"}, &bootstrapCluster); err != nil {
			return false, err
		}

		if bootstrapCluster.Spec.SkipClusterCreation && enableCreation {
			copy := bootstrapCluster.DeepCopy()
			copy.Spec.SkipClusterCreation = false
			if err := client.Update(context.Background(), copy); err != nil {
				return false, err
			}
			return false, nil
		}

		if bootstrapCluster.Status.ProviderStatus == nil {
			return false, nil
		}

		if !bootstrapCluster.Status.ProviderStatus.Ready {
			if bootstrapCluster.Status.ProviderStatus.Phase == bv1alpha1.Error {
				errorCount++
			}
			if errorCount == 10 {
				return false, fmt.Errorf("\n %s", bootstrapCluster.Status.ProviderStatus.Message)
			}
			return false, nil
		} else if !providerReady {
			errorCount = 0
			providerReady = true
			utils.Success("[1/5] Provider initialized successfully \n")
			utils.Warn("Waiting for CAPI operator ")
		}
		if !bootstrapCluster.Status.CapiOperatorStatus.Ready {
			utils.Warn(".")
			if bootstrapCluster.Status.CapiOperatorStatus.Phase == bv1alpha1.Error {
				errorCount++
			}
			if errorCount == 10 {
				return false, fmt.Errorf("\n %s", bootstrapCluster.Status.CapiOperatorStatus.Message)
			}
			return false, nil
		} else if !capiOperatorReady {
			errorCount = 0
			capiOperatorReady = true
			utils.Success("\n")
			utils.Success("[2/5] CAPI operator installed successfully \n")
			utils.Warn("Waiting for CAPI operator components ")

		}
		if !bootstrapCluster.Status.CapiOperatorComponentsStatus.Ready {
			utils.Warn(".")
			if bootstrapCluster.Status.CapiOperatorComponentsStatus.Phase == bv1alpha1.Error {
				errorCount++
			}
			if errorCount == 10 {
				return false, fmt.Errorf("\n %s", bootstrapCluster.Status.CapiOperatorComponentsStatus.Message)
			}
		} else if !capiOperatorComponentsReady {
			errorCount = 0
			capiOperatorComponentsReady = true
			utils.Success("\n")
			utils.Success("[3/5] CAPI operator components installed successfully \n")
			if !enableCreation {
				return true, nil
			}
			utils.Warn("Waiting for cluster ")
		}

		if !bootstrapCluster.Status.CapiClusterStatus.Ready {
			utils.Warn(".")
			if bootstrapCluster.Status.CapiClusterStatus.Phase == bv1alpha1.Error {
				errorCount++
			}
			if errorCount == 10 {
				return false, fmt.Errorf("\n %s", bootstrapCluster.Status.CapiClusterStatus.Message)
			}
		} else if !capiCluster {
			errorCount = 0
			capiCluster = true
			utils.Success("\n")
			utils.Success("[4/5] Cluster installed successfully \n")
			utils.Warn("Moving CAPI objects to the new cluster ")
		}
		if !bootstrapCluster.Status.Ready {
			utils.Warn(".")
			if bootstrapCluster.Status.Phase == bv1alpha1.Error {
				errorCount++
			}
			if errorCount == 10 {
				return false, fmt.Errorf("\n %s", bootstrapCluster.Status.Message)
			}
		} else if !moveReady {
			utils.Success("\n")
			utils.Success("[5/5] Moving cluster object to the new cluster finished successfully \n")
			return true, nil
		}

		return false, nil
	})
}

func (p *Plural) handleCreateNamespace(c *cli.Context) error {
	name := c.Args().Get(0)
	skipCreation := c.Bool("skip-if-exists")
	fmt.Printf("Creating namespace %s ...\n", name)
	err := p.InitKube()
	if err != nil {
		return err
	}
	if err := p.CreateNamespace(name); err != nil {
		if apierrors.IsAlreadyExists(err) && skipCreation {
			return nil
		}
		return err
	}

	return nil
}

func handleDeleteCluster(c *cli.Context) error {
	name := c.Args().Get(0)
	provider := cluster.NewProvider()
	fmt.Printf("Deleting cluster %s ...\n", name)
	return provider.Delete(name, "")
}

func handleCreateCluster(c *cli.Context) error {
	name := c.Args().Get(0)
	imageFlag := c.String("image")
	skipCreation := c.Bool("skip-if-exists")
	provider := cluster.NewProvider()
	fmt.Printf("Creating cluster %s ...\n", name)
	n, err := provider.ListNodes(name)
	if err != nil {
		return err
	}
	if len(n) != 0 && skipCreation {
		fmt.Printf("Cluster %s already exists \n", name)
		return nil
	}
	if err := provider.Create(
		name,
		cluster.CreateWithNodeImage(imageFlag),
		cluster.CreateWithRetain(false),
		cluster.CreateWithDisplayUsage(true),
		cluster.CreateWithDisplaySalutation(true),
	); err != nil {
		return errors.Wrap(err, "failed to create cluster")
	}
	kubeconfig, err := provider.KubeConfig(name, false)
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

	internalKubeconfig, err := provider.KubeConfig(name, true)
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

func WaitFor(timeout, interval time.Duration, f func() (bool, error)) error {
	var lastErr string
	timeup := time.After(timeout)
	for {
		select {
		case <-timeup:
			return fmt.Errorf("Time limit exceeded. Last error: %s", lastErr)
		default:
		}

		stop, err := f()
		if stop {
			return nil
		}
		if err != nil {
			return err
		}

		time.Sleep(interval)
	}
}

func getCRDList(ctx context.Context, client ctrlruntime.Client, crdList *apiextensionsv1.CustomResourceDefinitionList) error {
	if err := client.List(ctx, crdList); err != nil {
		return err
	}
	return nil
}
