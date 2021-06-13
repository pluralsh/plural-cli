package bundle

import (
	"fmt"
	"github.com/pluralsh/plural/pkg/api"
	"github.com/pluralsh/plural/pkg/utils"
	"strconv"
)

func configure(ctx map[string]interface{}, item *api.ConfigurationItem) error {
	res, err := fetchResult(ctx, item)
	if err != nil {
		return err
	}

	switch (item.Type) {
	case Int:
		parsed, err := strconv.Atoi(res)
		if err != nil {
			return err
		}
		ctx[item.Name] = parsed
	case Bool:
		parsed, err := strconv.ParseBool(res)
		if err != nil {
			return err
		}
		ctx[item.Name] = parsed
	case String:
		ctx[item.Name] = res
	}

	return nil
}

func fetchResult(ctx map[string]interface{}, item *api.ConfigurationItem) (string, error) {
	utils.Highlight(item.Name)
	fmt.Printf("\n>> %s\n", item.Documentation)

	def := item.Default
	prev, ok := ctx[item.Name]
	if ok {
		def = utils.ToString(prev)
	}

	if def != "" {
		return utils.ReadLineDefault("Enter the value", def)
	}

	return utils.ReadLine("Enter the value")
}
