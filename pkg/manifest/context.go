package manifest

import (
	"fmt"
	"os"
	"path/filepath"

	jsoniter "github.com/json-iterator/go"
	"gopkg.in/yaml.v2"

	"github.com/pluralsh/plural/pkg/api"
)

type Bundle struct {
	Repository string `json:"repository"`
	Name       string `json:"name"`
}

type SMTP struct {
	Service  string
	Server   string
	Port     int
	Sender   string
	User     string
	Password string
}

type Globals struct {
	CertIssuer   string `yaml:"certIssuer"`
	IngressClass string `yaml:"ingressClass"`
}

type Context struct {
	Bundles       []*Bundle
	Buckets       []string
	Domains       []string
	Protect       []string `yaml:"protect,omitempty" json:"protect,omitempty"`
	SMTP          *SMTP    `yaml:"smtp,omitempty"`
	Globals       *Globals `yaml:"globals,omitempty" json:"globals,omitempty"`
	Configuration map[string]map[string]interface{}
}

func (this *Context) MarshalJSON() ([]byte, error) {
	json := jsoniter.ConfigCompatibleWithStandardLibrary

	return json.Marshal(&struct {
		Bundles       []*Bundle                         `json:"bundles"`
		Buckets       []string                          `json:"buckets"`
		Domains       []string                          `json:"domains"`
		Protect       []string                          `yaml:"protect,omitempty" json:"protect,omitempty"`
		SMTP          *SMTP                             `yaml:"smtp,omitempty" json:"smtp"`
		Globals       *Globals                          `yaml:"globals,omitempty" json:"globals,omitempty"`
		Configuration map[string]map[string]interface{} `json:"configuration"`
	}{
		Bundles:       this.Bundles,
		Buckets:       this.Buckets,
		Domains:       this.Domains,
		Protect:       this.Protect,
		SMTP:          this.SMTP,
		Globals:       this.Globals,
		Configuration: this.Configuration,
	})
}

type VersionedContext struct {
	ApiVersion string   `yaml:"apiVersion"`
	Kind       string   `yaml:"kind"`
	Spec       *Context `yaml:"spec"`
}

func ContextPath() string {
	path, _ := filepath.Abs("context.yaml")
	return path
}

func BuildContext(path string, insts []*api.Installation) error {
	ctx := &Context{
		Configuration: make(map[string]map[string]interface{}),
	}

	for _, inst := range insts {
		ctx.Configuration[inst.Repository.Name] = inst.Context
	}

	return ctx.Write(path)
}

func FetchContext() (*Context, error) {
	return ReadContext(ContextPath())
}

func ReadContext(path string) (c *Context, err error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return
	}

	ctx := &VersionedContext{}
	err = yaml.Unmarshal(contents, ctx)
	c = ctx.Spec
	return
}

func NewContext() *Context {
	return &Context{
		Bundles: make([]*Bundle, 0),
		// Globals:       &Globals{CertIssuer: "plural"},
		Configuration: make(map[string]map[string]interface{}),
	}
}

func (c *Context) Repo(name string) (res map[string]interface{}, ok bool) {
	res, ok = c.Configuration[name]
	return
}

func (c *Context) AddBundle(repo, name string) {
	for _, b := range c.Bundles {
		if b.Name == name && b.Repository == repo {
			return
		}
	}

	c.Bundles = append(c.Bundles, &Bundle{Repository: repo, Name: name})
}

func (c *Context) AddBucket(bucket string) {
	c.Buckets = append(c.Buckets, bucket)
}

func (c *Context) HasBucket(bucket string) bool {
	for _, b := range c.Buckets {
		if b == bucket {
			return true
		}
	}

	return false
}

func (c *Context) AddDomain(domain string) {
	c.Domains = append(c.Domains, domain)
}

func (c *Context) Protected(name string) bool {
	for _, r := range c.Protect {
		if r == name {
			return true
		}
	}

	return false
}

func (c *Context) HasDomain(domain string) bool {
	// Exclusion for empty string.
	// There are some cases where an empty string for the hostname is used.
	if domain == "" {
		return false
	}
	for _, d := range c.Domains {
		if d == domain {
			return true
		}
	}

	return false
}

func (c *Context) Write(path string) error {
	versioned := &VersionedContext{
		ApiVersion: "plural.sh/v1alpha1",
		Kind:       "Context",
		Spec:       c,
	}

	io, err := yaml.Marshal(versioned)
	if err != nil {
		return err
	}

	return os.WriteFile(path, io, 0644)
}

func (c *Context) ContainsString(str, msg, ignoreRepo, ignoreKey string) error {
	for r, section := range c.Configuration {
		for k, val := range section {
			if v, ok := val.(string); ok && v == str && (r != ignoreRepo || k != ignoreKey) {
				return fmt.Errorf(msg)
			}
		}
	}

	return nil
}

func (smtp *SMTP) GetServer() string {
	if smtp.Service != "" {
		if val, ok := smtpConfig[smtp.Service]; ok {
			return val.Server
		}
	}
	return smtp.Server
}

func (smtp *SMTP) GetPort() int {
	if smtp.Service != "" {
		if val, ok := smtpConfig[smtp.Service]; ok {
			return val.Port
		}
	}
	return smtp.Port
}

func (smtp *SMTP) Configuration() map[string]interface{} {
	return map[string]interface{}{
		"Server":   smtp.GetServer(),
		"Port":     smtp.GetPort(),
		"User":     smtp.User,
		"Password": smtp.Password,
		"Service":  smtp.Service,
		"Sender":   smtp.Sender,
	}
}
