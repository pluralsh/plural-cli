package manifest

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/pluralsh/plural/pkg/api"
	"gopkg.in/yaml.v2"

	"k8s.io/apimachinery/pkg/util/sets"
)

type Bundle struct {
	Repository string
	Name       string
}

type SMTP struct {
	Service  string
	Server   string
	Port     int
	Sender   string
	User     string
	Password string
}

type Context struct {
	Bundles       []*Bundle
	Buckets       sets.String
	Domains       sets.String
	SMTP          *SMTP `yaml:"smtp,omitempty"`
	Configuration map[string]map[string]interface{}
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

func ReadContext(path string) (c *Context, err error) {
	contents, err := ioutil.ReadFile(path)
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
		Bundles:       make([]*Bundle, 0),
		Configuration: make(map[string]map[string]interface{}),
		Domains:       sets.NewString(),
		Buckets:       sets.NewString(),
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
	c.Buckets.Insert(bucket)
}

func (c *Context) HasBucket(bucket string) bool {
	return c.Buckets.Has(bucket)
}

func (c *Context) AddDomain(bucket string) {
	c.Domains.Insert(bucket)
}

func (c *Context) HasDomain(bucket string) bool {
	return c.Domains.Has(bucket)
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

	return ioutil.WriteFile(path, io, 0644)
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
