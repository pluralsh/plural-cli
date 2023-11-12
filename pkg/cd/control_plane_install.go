package cd

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/AlecAivazis/survey/v2"
	"github.com/osteele/liquid"
	"github.com/pluralsh/plural/pkg/api"
	"github.com/pluralsh/plural/pkg/bundle"
	"github.com/pluralsh/plural/pkg/config"
	"github.com/pluralsh/plural/pkg/crypto"
	"github.com/pluralsh/plural/pkg/utils"
)

var (
	liquidEngine = liquid.NewEngine()
)

const (
	templateUrl = "https://raw.githubusercontent.com/pluralsh/console/cd-scaffolding/charts/console/values.yaml.liquid"
)

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

	randoms := map[string]string{}
	for _, key := range []string{"jwt", "erlang", "adminPassword", "kasApi", "kasPrivateApi", "kasRedis"} {
		rand, err := crypto.RandStr(32)
		if err != nil {
			return "", err
		}
		randoms[key] = rand
	}

	configuration := map[string]string{
		"consoleDns":  resp.Console,
		"kasDns":      resp.KubeProxy,
		"aesKey":      utils.GenAESKey(),
		"adminName":   me.Email,
		"adminEmail":  me.Email,
		"clusterName": resp.ClusterName,
		"pluralToken": conf.Token,
		"postgresUrl": resp.PostgresDsn,
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

	tpl, err := fetchTemplate()
	if err != nil {
		return "", err
	}

	bindings := map[string]interface{}{
		"configuration": configuration,
	}

	res, err := liquidEngine.ParseAndRender(tpl, bindings)
	return string(res), err
}

func fetchTemplate() (res []byte, err error) {
	resp, err := http.Get(templateUrl)
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
