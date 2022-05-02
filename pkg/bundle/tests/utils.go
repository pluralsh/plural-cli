package tests

import (
	"github.com/pluralsh/plural/pkg/api"
	"github.com/pluralsh/plural/pkg/manifest"
)

type ContextValue struct {
	Val     interface{}
	Present bool
}

func collectArguments(args []*api.TestArgument, ctx *manifest.Context) map[string]*ContextValue {
	res := make(map[string]*ContextValue)
	for _, arg := range args {
		val, ok := ctx.Configuration[arg.Repo][arg.Key]
		res[arg.Name] = &ContextValue{val, ok}
	}
	return res
}
