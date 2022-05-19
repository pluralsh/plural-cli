package uiOld

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

	pages := tview.NewPages()
	// providerForm := &ProviderForm{}k

	providerForm := tview.NewForm()

	pages.AddPage("ProviderForm", providerForm, true, true)

	// AddButton("Save", nextSlide).
	// AddButton("Cancel", nextSlide)
	providerForm.SetBorder(true).SetTitle("Provider Setup")

	providerForm.AddDropDown("Provider", []string{AWS, GCP, AZURE, EQUINIX, KIND}, -1, func(option string, optionIndex int) { buildProviderForm(c, pages, providerForm, option, optionIndex) })

	// init := tview.NewFlex().
	// 	SetDirection(tview.FlexRow).
	// 	AddItem(pages, 0, 1, true)

	return "Workspace Config", pages
}

func buildProviderForm(c *cli.Context, pages *tview.Pages, providerForm *tview.Form, option string, optionIndex int) {

	resetFormInputs(providerForm)

	switch option {
	case AWS:
		providerForm.AddInputField("Cluster Name", "", 20, nil, nil).
			AddDropDown("Region", provider.AwsRegions, 18, nil)
		setupProvider(c, providerForm, pages)
	case GCP:
		providerForm.AddInputField("Cluster Name", "", 20, nil, nil).
			AddDropDown("Region", provider.GcpRegions, 12, nil).
			AddInputField("Project", "", 20, nil, nil)
		setupProvider(c, providerForm, pages)
	case AZURE:
		providerForm.AddInputField("Cluster Name", "", 20, nil, nil).
			AddInputField("Storage Account", "", 20, nil, nil).
			AddInputField("Resource Group", "", 20, nil, nil).
			AddDropDown("Region", provider.AzureRegions, 0, nil)
		setupProvider(c, providerForm, pages)
	case EQUINIX:
		providerForm.AddInputField("Cluster Name", "", 20, nil, nil).
			AddInputField("Facility", "", 20, nil, nil).
			AddInputField("Project Name", "", 20, nil, nil).
			AddPasswordField("API Token", "", 20, '*', nil)
		setupProvider(c, providerForm, pages)
	case KIND:
		providerForm.AddInputField("Cluster Name", "", 20, nil, nil)
		setupProvider(c, providerForm, pages)
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
	providerForm.ClearButtons()
}

func setupProvider(c *cli.Context, providerForm *tview.Form, pages *tview.Pages) {
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
				mkAws(conf, providerForm, pages)
			}
		}).
		AddButton("Cancel", func() {
			app.Stop()
		}).
		AddButton("Error-test", func() {
			pages.AddAndSwitchToPage("error", ErrorModal(pages, fmt.Errorf("test-error"), 40, 10), true)
		})
}

func mkAws(conf config.Config, providerForm *tview.Form, pages *tview.Pages) {

	cluster := providerForm.GetFormItemByLabel("Cluster Name").(*tview.InputField).GetText()
	_, region := providerForm.GetFormItemByLabel("Region").(*tview.DropDown).GetCurrentOption()
	bucketPrefix := providerForm.GetFormItemByLabel("Bucket Prefix").(*tview.InputField).GetText()
	subdomain := providerForm.GetFormItemByLabel("Domain name").(*tview.InputField).GetText()
	pluralDns := providerForm.GetFormItemByLabel("Use Plural DNS").(*tview.Checkbox).IsChecked()
	providerStruct := &provider.AWSProvider{}
	providerStruct.Clus = cluster
	providerStruct.Reg = region

	client, err := getClient(providerStruct.Reg)
	if err != nil {
		pages.AddAndSwitchToPage("error", ErrorModal(pages, err, 40, 10), true)
	}

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

func ErrorModal(pages *tview.Pages, err error, width, height int) (content tview.Primitive) {
	// modal := func(p tview.Primitive, width, height int) string, tview.Primitive {
	// 	return "Modal", tview.NewFlex().
	// 		AddItem(nil, 0, 1, false).
	// 		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
	// 			AddItem(nil, 0, 1, false).
	// 			AddItem(p, height, 1, false).
	// 			AddItem(nil, 0, 1, false), width, 1, false).
	// 		AddItem(nil, 0, 1, false)
	// }
	modal := tview.NewModal().
		SetText(fmt.Sprint(err)).
		AddButtons([]string{"Close"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "Close" {
				pages.RemovePage("error")
			}
		})

	// errorView := tview.NewTextView().
	// 	SetWrap(false).
	// 	SetDynamicColors(true).SetTitle("Error")
	// errorView.SetBorderPadding(1, 1, 2, 0).SetBorder(true)
	// fmt.Fprint(errorView, err)

	// return tview.NewFlex().
	// 	AddItem(Center(width, height, p), 0, 1, true).
	// 	AddItem(errorView, codeWidth, 1, false)

	// box := tview.NewBox().
	// 	SetBorder(true).
	// 	SetTitle("Centered Box")

	return tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(modal, 20, 1, false).
			AddItem(nil, 0, 1, false), 20, 1, false).
		AddItem(nil, 0, 1, false)
}
