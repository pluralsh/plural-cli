package helm

import (
	"os"
	"fmt"
	"io/ioutil"
	"strings"
	"github.com/chartmuseum/helm-push/pkg/helm"
	"github.com/pluralsh/plural/pkg/config"
	cm "github.com/chartmuseum/helm-push/pkg/chartmuseum"
)

func Push(chartName, repoUrl string) error {
	repo, err := helm.TempRepoFromURL(repoUrl)
	if err != nil {
		return err
	}

	chart, err := helm.GetChartByName(chartName)
	if err != nil {
		return err
	}

	conf := config.Read()

	url := strings.Replace(repo.Config.URL, "cm://", "https://", 1)
	client, err := cm.NewClient(
		cm.URL(url),
		cm.AccessToken(conf.Token),
		cm.ContextPath("/cm"),
	)

	tmp, err := ioutil.TempDir("", "helm-push-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)

	chartPackagePath, err := helm.CreateChartPackage(chart, tmp)
	if err != nil {
		return err
	}

	resp, err := client.UploadChartPackage(chartPackagePath, false)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 && resp.StatusCode != 202 {
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("Failed to upload to plural, code %d error %s\n", resp.StatusCode, string(b))
	}
	fmt.Println("Done.")
	return nil
}