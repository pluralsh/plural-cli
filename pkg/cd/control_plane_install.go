package cd

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/AlecAivazis/survey/v2"
	pollytemplate "github.com/pluralsh/console/go/polly/template"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/client-go/dynamic"
	k8sclient "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"

	"github.com/pluralsh/plural-cli/pkg/api"
	"github.com/pluralsh/plural-cli/pkg/bundle"
	"github.com/pluralsh/plural-cli/pkg/config"
	"github.com/pluralsh/plural-cli/pkg/crypto"
	"github.com/pluralsh/plural-cli/pkg/kubernetes"
	"github.com/pluralsh/plural-cli/pkg/manifest"
	"github.com/pluralsh/plural-cli/pkg/provider"
	"github.com/pluralsh/plural-cli/pkg/template"
	"github.com/pluralsh/plural-cli/pkg/utils"
	"github.com/pluralsh/plural-cli/pkg/utils/git"
)

const (
	templateUrl = "https://raw.githubusercontent.com/pluralsh/console/master/templates/values.yaml.liquid"
	tplUrl      = "https://raw.githubusercontent.com/pluralsh/console/master/templates/values.yaml.tpl"

	noClusterIssuer     = "(none)"
	defaultIngressClass = "nginx"
)

// discoveryTimeout bounds cluster lookups so an unresponsive API server or credential plugin falls back to manual input.
var discoveryTimeout = 5 * time.Second

type secrets struct {
	AesKey string `yaml:"aes_key"`
	Erlang string `yaml:"erlang"`
}

type ingress struct {
	ConsoleDns string `yaml:"console_dns"`
	KasDns     string `yaml:"kas_dns"`
}

type consoleValues struct {
	Ingress ingress `yaml:"ingress"`
	Secrets secrets `yaml:"secrets"`
}

// ObjectStoreOpts holds cloud-specific console object-store settings for
// control-plane values. Only one provider's fields should be set.
type ObjectStoreOpts struct {
	S3Bucket          string
	S3Region          string
	AzureAccount      string
	AzureContainer    string
	AzureClientId     string
	GcsBucket         string
	GcsServiceAccount string
}

func ControlPlaneValues(conf config.Config, file, domain, dsn, name, clusterIssuer, ingressClass string, objectStore ObjectStoreOpts) (string, error) {
	consoleDns := fmt.Sprintf("console.%s", domain)
	kasDns := fmt.Sprintf("kas.%s", domain)
	existing := consoleValues{}
	if utils.Exists(file) {
		if d, err := utils.ReadFile(file); err == nil {
			if err := yaml.Unmarshal([]byte(d), &existing); err == nil {
				if existing.Ingress.ConsoleDns != "" {
					consoleDns = existing.Ingress.ConsoleDns
				}
				if existing.Ingress.KasDns != "" {
					kasDns = existing.Ingress.KasDns
				}
			}
		}
	}

	randoms := map[string]string{}
	randoms["erlang"] = crypto.RandString(64)
	for _, key := range []string{"jwt", "adminPassword", "kasApi", "kasPrivateApi", "kasRedis"} {
		rand, err := crypto.RandStr(32)
		if err != nil {
			return "", err
		}
		randoms[key] = rand
	}

	if existing.Secrets.Erlang != "" {
		randoms["erlang"] = existing.Secrets.Erlang
	}

	client := api.FromConfig(&conf)
	me, err := client.Me()
	if err != nil {
		return "", fmt.Errorf("you must run `plural login` before installing")
	}

	root, err := git.Root()
	if err != nil {
		return "", err
	}

	project, err := manifest.ReadProject(filepath.Join(root, "workspace.yaml"))
	if err != nil {
		return "", err
	}

	prov, err := provider.FromManifest(project)
	if err != nil {
		return "", err
	}

	configuration := map[string]interface{}{
		"consoleDns":    consoleDns,
		"kasDns":        kasDns,
		"aesKey":        utils.GenAESKey(),
		"adminName":     me.Email,
		"adminEmail":    me.Email,
		"clusterName":   name,
		"pluralToken":   conf.Token,
		"postgresUrl":   dsn,
		"provider":      prov.Name(),
		"clusterIssuer": clusterIssuer,
	}

	if objectStore.S3Bucket != "" {
		configuration["objectStoreBucket"] = objectStore.S3Bucket
		configuration["objectStoreRegion"] = objectStore.S3Region
	}
	if objectStore.AzureAccount != "" {
		configuration["objectStoreAzureStorageAccount"] = objectStore.AzureAccount
		configuration["objectStoreAzureContainer"] = objectStore.AzureContainer
		if objectStore.AzureClientId != "" {
			configuration["objectStoreAzureClientId"] = objectStore.AzureClientId
		}
	}
	if objectStore.GcsBucket != "" {
		configuration["objectStoreGcsBucket"] = objectStore.GcsBucket
		if objectStore.GcsServiceAccount != "" {
			configuration["objectStoreGcsServiceAccount"] = objectStore.GcsServiceAccount
		}
	}

	for _, val := range []string{clusterIssuer, ingressClass} {
		if err := validateResourceName(val); err != nil {
			return "", err
		}
	}
	if ingressClass == "" {
		ingressClass = contextIngressClass(prov.Context())
	}
	if ingressClass != "" {
		configuration["ingressClass"] = ingressClass
	}

	if existing.Secrets.AesKey != "" {
		configuration["aesKey"] = existing.Secrets.AesKey
	}

	for k, v := range randoms {
		configuration[k] = v
	}

	if cryptos, err := cryptoVals(); err == nil {
		for k, v := range cryptos {
			configuration[k] = v
		}
	}

	clientId, clientSecret, err := ensureInstalledAndOidc(client, consoleDns)
	if err != nil {
		return "", err
	}
	configuration["pluralClientId"] = clientId
	configuration["pluralClientSecret"] = clientSecret

	tpl, err := fetchTemplate(tplUrl)
	if err != nil {
		return "", err
	}

	return template.RenderString(string(tpl), configuration)
}

func cryptoVals() (map[string]string, error) {
	res := make(map[string]string)
	keyFile, err := config.PluralDir("key")
	if err != nil {
		return res, err
	}

	aes, err := utils.ReadFile(keyFile)
	if err != nil {
		return res, err
	}
	res["key"] = aes

	identityFile, err := config.PluralDir("identity")
	if err != nil {
		return res, nil
	}

	identity, err := utils.ReadFile(identityFile)
	if err != nil {
		return res, nil
	}
	res["identity"] = identity
	return res, nil
}

func CreateControlPlane(conf config.Config) (string, error) {
	client := api.FromConfig(&conf)
	me, err := client.Me()
	if err != nil {
		return "", fmt.Errorf("you must run `plural login` before installing")
	}

	azureSurvey := []*survey.Question{
		{
			Name:   "console",
			Prompt: &survey.Input{Message: "Enter a dns name for your installation of the console (eg console.your.domain):"},
		},
		{
			Name:   "kubeProxy",
			Prompt: &survey.Input{Message: "Enter a dns name for the kube proxy (eg kas.your.domain), this is used for dashboarding functionality:"},
		},
		{
			Name:   "clusterName",
			Prompt: &survey.Input{Message: "Enter a name for this cluster:"},
		},
		{
			Name:   "postgresDsn",
			Prompt: &survey.Input{Message: "Enter a postgres connection string for the underlying database (should be postgres://<user>:<password>@<host>:5432/<database>):"},
		},
	}
	var resp struct {
		Console     string
		KubeProxy   string
		ClusterName string
		PostgresDsn string
	}
	if err := survey.Ask(azureSurvey, &resp); err != nil {
		return "", err
	}

	clusterIssuer, ingressClass, err := askIngressSettings()
	if err != nil {
		return "", err
	}

	randoms := map[string]string{}
	randoms["erlang"] = crypto.RandString(64)
	for _, key := range []string{"jwt", "adminPassword", "kasApi", "kasPrivateApi", "kasRedis"} {
		rand, err := crypto.RandStr(32)
		if err != nil {
			return "", err
		}
		randoms[key] = rand
	}

	configuration := map[string]string{
		"consoleDns":   resp.Console,
		"kasDns":       resp.KubeProxy,
		"aesKey":       utils.GenAESKey(),
		"adminName":    me.Email,
		"adminEmail":   me.Email,
		"clusterName":  resp.ClusterName,
		"pluralToken":  conf.Token,
		"postgresUrl":  resp.PostgresDsn,
		"ingressClass": ingressClass,
	}
	// liquid treats "" as truthy, so only set the issuer when one was picked
	if clusterIssuer != "" {
		configuration["clusterIssuer"] = clusterIssuer
	}
	for k, v := range randoms {
		configuration[k] = v
	}

	clientId, clientSecret, err := ensureInstalledAndOidc(client, resp.Console)
	if err != nil {
		return "", err
	}
	configuration["pluralClientId"] = clientId
	configuration["pluralClientSecret"] = clientSecret

	tpl, err := fetchTemplate(templateUrl)
	if err != nil {
		return "", err
	}

	bindings := map[string]interface{}{
		"configuration": configuration,
	}

	res, err := pollytemplate.RenderLiquid(tpl, bindings)
	return string(res), err
}

// contextIngressClass reads IngressClass from the provider context; utils.ToString would turn a missing key into "<nil>".
func contextIngressClass(ctx map[string]interface{}) string {
	ingressClass, _ := ctx["IngressClass"].(string)
	return ingressClass
}

// askIngressSettings lets the user pick a cluster issuer and ingress class from the current cluster,
// falling back to free text when they can't be listed.
func askIngressSettings() (clusterIssuer, ingressClass string, err error) {
	var issuers, classes []string
	if kubeConf, err := kubernetes.KubeConfig(); err == nil {
		issuers, classes = listIngressOptions(kubeConf)
	}

	issuerMsg := "Enter the cert-manager cluster issuer for the console ingresses (leave blank for none):"
	if len(issuers) > 0 {
		// first option is the select's default, so Enter means no issuer
		issuers = append([]string{noClusterIssuer}, issuers...)
		issuerMsg = "Select the cert-manager cluster issuer for the console ingresses:"
	}
	clusterIssuer, err = selectOrInput(issuerMsg, issuers, "")
	if err != nil {
		return
	}
	if clusterIssuer == noClusterIssuer {
		clusterIssuer = ""
	}

	classMsg := "Enter the ingress class for the console ingresses:"
	if len(classes) > 0 {
		classMsg = "Select the ingress class for the console ingresses:"
	}
	ingressClass, err = selectOrInput(classMsg, classes, defaultIngressClass)
	return
}

// listIngressOptions lists cluster issuers and ingress classes, returning empty lists on any error or after discoveryTimeout.
func listIngressOptions(kubeConf *rest.Config) ([]string, []string) {
	conf := rest.CopyConfig(kubeConf)
	if conf.ExecProvider != nil {
		// CopyConfig shares ExecProvider; copy it so credential plugins can't prompt over the survey UI
		exec := *conf.ExecProvider
		exec.InteractiveMode = clientcmdapi.NeverExecInteractiveMode
		conf.ExecProvider = &exec
	}

	// shortcut: on timeout the goroutine is abandoned until the CLI exits, fine for a one-shot command
	var issuers, classes []string
	done := make(chan struct{})
	go func() {
		defer close(done)
		if dyn, err := dynamic.NewForConfig(conf); err == nil {
			issuers, _ = kubernetes.ClusterIssuers(dyn)
		}
		if cs, err := k8sclient.NewForConfig(conf); err == nil {
			classes, _ = kubernetes.IngressClasses(cs)
		}
	}()

	select {
	case <-done:
		return issuers, classes
	case <-time.After(discoveryTimeout):
		return nil, nil
	}
}

func selectOrInput(message string, options []string, def string) (res string, err error) {
	if len(options) > 0 {
		err = survey.AskOne(&survey.Select{Message: message, Options: options}, &res)
		return
	}
	// only free text needs checking; select options come from the cluster
	err = survey.AskOne(&survey.Input{Message: message, Default: def}, &res, survey.WithValidator(func(ans interface{}) error {
		return validateResourceName(utils.ToString(ans))
	}))
	return
}

// validateResourceName accepts an empty value or a valid Kubernetes resource name, since these values are rendered into YAML as-is.
func validateResourceName(val string) error {
	if val == "" {
		return nil
	}
	if errs := validation.IsDNS1123Subdomain(val); len(errs) > 0 {
		return fmt.Errorf("%q is not a valid Kubernetes resource name: %s", val, strings.Join(errs, ", "))
	}
	return nil
}

func fetchTemplate(url string) (res []byte, err error) {
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	var out bytes.Buffer
	_, err = io.Copy(&out, resp.Body)
	return out.Bytes(), err
}

func ensureInstalledAndOidc(client api.Client, dns string) (clientId string, clientSecret string, err error) {
	inst, err := client.GetInstallation("console")
	if err != nil || inst == nil {
		repo, err := client.GetRepository("console")
		if err != nil {
			return "", "", err
		}
		_, err = client.CreateInstallation(repo.Id)
		if err != nil {
			return "", "", err
		}
	}

	redirectUris := []string{fmt.Sprintf("https://%s/oauth/callback", dns)}
	err = bundle.SetupOIDC("console", client, redirectUris, "POST")
	if err != nil {
		return
	}

	inst, err = client.GetInstallation("console")
	if err != nil {
		return
	}

	return inst.OIDCProvider.ClientId, inst.OIDCProvider.ClientSecret, nil
}
