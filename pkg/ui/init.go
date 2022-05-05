package ui

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/pluralsh/plural/pkg/config"
	"github.com/pluralsh/plural/pkg/manifest"
	"github.com/pluralsh/plural/pkg/provider"
	"github.com/pluralsh/plural/pkg/utils/errors"
	"github.com/rivo/tview"
	"github.com/urfave/cli"
)

const (
	GCP     = provider.GCP
	AWS     = provider.AWS
	AZURE   = provider.AZURE
	EQUINIX = provider.EQUINIX
	KIND    = provider.KIND
)

// type ProviderForm struct {
// 	form *tview.Form
// }

func Init(c *cli.Context, nextSlide func()) (title string, content tview.Primitive) {

	// providerForm := &ProviderForm{}k

	providerForm := tview.NewForm()
	// AddButton("Save", nextSlide).
	// AddButton("Cancel", nextSlide)
	providerForm.SetBorder(true).SetTitle("Provider Setup")

	providerForm.AddDropDown("Provider", []string{AWS, GCP, AZURE, EQUINIX, KIND}, 0, func(option string, optionIndex int) { buildProviderForm(c, providerForm, option, optionIndex) })

	return "Workspace Config", providerForm
}

func buildProviderForm(c *cli.Context, providerForm *tview.Form, option string, optionIndex int) {

	resetFormInputs(providerForm)

	switch option {
	case AWS:
		providerForm.AddInputField("Cluster Name", "", 20, nil, nil).
			AddDropDown("Region", provider.AwsRegions, 18, nil)
		setupProvider(c, providerForm)
	case GCP:
		providerForm.AddInputField("Cluster Name", "", 20, nil, nil).
			AddDropDown("Region", provider.GcpRegions, 12, nil).
			AddInputField("Project", "", 20, nil, nil)
		setupProvider(c, providerForm)
	case AZURE:
		providerForm.AddInputField("Cluster Name", "", 20, nil, nil).
			AddInputField("Storage Account", "", 20, nil, nil).
			AddInputField("Resource Group", "", 20, nil, nil).
			AddDropDown("Region", provider.AzureRegions, 0, nil)
		setupProvider(c, providerForm)
	case EQUINIX:
		providerForm.AddInputField("Cluster Name", "", 20, nil, nil).
			AddInputField("Facility", "", 20, nil, nil).
			AddInputField("Project Name", "", 20, nil, nil).
			AddPasswordField("API Token", "", 20, '*', nil)
		setupProvider(c, providerForm)
	case KIND:
		providerForm.AddInputField("Cluster Name", "", 20, nil, nil)
		setupProvider(c, providerForm)
	}
}

// Reset the form values in case the provider selection changes
func resetFormInputs(providerForm *tview.Form) {
	inputCount := providerForm.GetFormItemCount()
	if inputCount > 1 {
		for i := inputCount - 1; i >= 1; i-- {
			providerForm.RemoveFormItem(i)
		}
	}
}

func setupProvider(c *cli.Context, providerForm *tview.Form) {
	providerForm.AddInputField("Bucket Prefix", "", 20, nil, nil).
		AddCheckbox("Use Plural DNS", true, nil).
		AddInputField("Domain name", "", 20, nil, nil).
		AddButton("Save", func() {
			_, provider := providerForm.GetFormItemByLabel("Provider").(*tview.DropDown).GetCurrentOption()
			conf := config.Read()
			conf.Token = ""
			conf.Endpoint = c.String("endpoint") //TODO: not actuall functional
			switch provider {
			case AWS:
				mkAws(conf, providerForm)
			}
		}).
		AddButton("Cancel", func() {
			app.Stop()
		})
}

func mkAws(conf config.Config, providerForm *tview.Form) {

	cluster := providerForm.GetFormItemByLabel("Cluster Name").(*tview.InputField).GetText()
	_, region := providerForm.GetFormItemByLabel("Region").(*tview.DropDown).GetCurrentOption()
	bucketPrefix := providerForm.GetFormItemByLabel("Bucket Prefix").(*tview.InputField).GetText()
	subdomain := providerForm.GetFormItemByLabel("Domain name").(*tview.InputField).GetText()
	pluralDns := providerForm.GetFormItemByLabel("Use Plural DNS").(*tview.Checkbox).IsChecked()
	providerStruct := &provider.AWSProvider{}
	providerStruct.Clus = cluster
	providerStruct.Reg = region

	client, _ := getClient(providerStruct.Reg)
	// if err != nil {
	// 	return nil, err
	// }

	account, _ := provider.GetAwsAccount()
	// if err != nil {
	// 	return nil, errors.ErrorWrap(err, "Failed to get aws account (is your aws cli configured?)")
	// }

	// if len(account) <= 0 {
	// 	return nil, errors.ErrorWrap(fmt.Errorf("Unable to find aws account id, is your aws cli configured?"), "AWS cli error:")
	// }

	providerStruct.Proj = account
	providerStruct.StorageClient = client

	projectManifest := manifest.ProjectManifest{
		Cluster:      providerStruct.Cluster(),
		Project:      providerStruct.Project(),
		Provider:     AWS,
		Region:       providerStruct.Region(),
		Bucket:       fmt.Sprintf("%s-tf-state", bucketPrefix),
		BucketPrefix: bucketPrefix,
		Network: &manifest.NetworkConfig{
			Subdomain: subdomain,
			PluralDns: pluralDns,
		},
		Owner: &manifest.Owner{Email: conf.Email, Endpoint: conf.Endpoint},
	}
	projectManifest.Write(manifest.ProjectManifestPath())
}

func getClient(region string) (*s3.S3, error) {
	sess, err := session.NewSessionWithOptions(session.Options{
		Config:            aws.Config{Region: aws.String(region)},
		SharedConfigState: session.SharedConfigEnable,
	})

	if err != nil {
		return nil, errors.ErrorWrap(err, "Failed to initialize aws client: ")
	}

	return s3.New(sess), nil
}
