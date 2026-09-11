package up

import (
	"context"
	"fmt"
	"strings"

	"github.com/pluralsh/plural-cli/pkg/api"
	"github.com/pluralsh/plural-cli/pkg/common"
	"github.com/pluralsh/plural-cli/pkg/console"
)

// ConsoleInstance is one Plural Cloud Console from GetConsoleInstances.
type ConsoleInstance struct {
	ID   string
	Name string
	URL  string
}

// InstanceLister lists cloud Console instances (App GraphQL).
type InstanceLister interface {
	List(ctx context.Context) ([]ConsoleInstance, error)
}

// LiveInstanceLister uses the App API client (same plane as CLI InitPluralClient).
type LiveInstanceLister struct{}

// DefaultInstanceLister returns the live App GraphQL lister.
func DefaultInstanceLister() InstanceLister { return LiveInstanceLister{} }

// List returns Console instances for the logged-in Plural account.
func (LiveInstanceLister) List(ctx context.Context) ([]ConsoleInstance, error) {
	_ = ctx
	instances, err := api.NewClient().GetConsoleInstances()
	if err != nil {
		return nil, err
	}
	out := make([]ConsoleInstance, 0, len(instances))
	for _, inst := range instances {
		if inst == nil {
			continue
		}
		out = append(out, ConsoleInstance{ID: inst.ID, Name: inst.Name, URL: inst.URL})
	}
	return out, nil
}

// DefaultInstanceIndex returns the index matching prior console.yml hostname, or 0.
func DefaultInstanceIndex(instances []ConsoleInstance, priorURL string) int {
	if priorURL == "" || len(instances) == 0 {
		return 0
	}
	priorHost := common.GetHostnameFromURL(priorURL)
	for i, inst := range instances {
		if strings.EqualFold(priorHost, common.GetHostnameFromURL(inst.URL)) {
			return i
		}
	}
	return 0
}

// PriorConsoleConfig matches selected URL hostname (HandleCdLogin Affirm path).
func PriorConsoleMatches(priorURL, selectedURL string) bool {
	if priorURL == "" || selectedURL == "" {
		return false
	}
	return strings.EqualFold(common.GetHostnameFromURL(priorURL), common.GetHostnameFromURL(selectedURL))
}

// ReadPriorConsole returns console.yml Url/Token (may be empty).
func ReadPriorConsole() console.Config {
	return console.ReadConfig()
}

// SaveConsoleConfig writes console.yml after CD login (URL + token).
func SaveConsoleConfig(rawURL, token string) error {
	conf := console.Config{
		Url:   console.NormalizeUrl(rawURL),
		Token: strings.TrimSpace(token),
	}
	return conf.Save()
}

// ValidateConsoleConfig mirrors cmd/command/up ValidateConsoleConfig:
// console.yml must match one of the listed instances.
func ValidateConsoleConfig(instances []ConsoleInstance, conf console.Config) error {
	if conf.Url == "" {
		return fmt.Errorf("you haven't configured your Plural Console client yet")
	}
	var id string
	for _, inst := range instances {
		if strings.Contains(conf.Url, inst.URL) {
			id = inst.ID
			break
		}
		if strings.EqualFold(common.GetHostnameFromURL(conf.Url), common.GetHostnameFromURL(inst.URL)) {
			id = inst.ID
			break
		}
	}
	if id == "" {
		return fmt.Errorf("your configuration doesn't match to any existing Plural Console")
	}
	return nil
}
