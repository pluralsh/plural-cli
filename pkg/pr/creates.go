package pr

import (
	"path/filepath"

	"github.com/pluralsh/plural-cli/pkg/utils"
	"github.com/pluralsh/polly/algorithms"
)

type replacement struct {
	source string
	dest   string
}

func applyCreates(creates *CreateSpec, ctx map[string]interface{}) error {
	if creates == nil {
		return nil
	}

	for _, tpl := range creates.Templates {
		enabled, err := evaluateCondition(tpl.Condition, ctx)
		if err != nil {
			return err
		}

		if !enabled {
			continue
		}

		source := tpl.Source
		if len(tpl.Context) > 0 {
			ctx = algorithms.Merge(ctx, tpl.Context)
		}

		if tpl.External {
			source = filepath.Join(creates.ExternalDir, source)
		}
		if templatedSource, err := templateReplacement([]byte(source), ctx); err == nil {
			source = string(templatedSource)
		}

		destPath := []byte(tpl.Destination)
		dest, err := templateReplacement(destPath, ctx)
		if err != nil {
			dest = destPath
		}

		replacements := []replacement{{source, string(dest)}}
		if utils.IsDir(source) {
			files, err := utils.ListDirectory(source)
			if err != nil {
				return err
			}

			replacements = []replacement{}
			for _, file := range files {
				destFile, err := filepath.Rel(source, file)
				if err != nil {
					return err
				}
				destFile = filepath.Join(string(dest), destFile)
				replacements = append(replacements, replacement{source: file, dest: destFile})
			}
		}

		for _, replacement := range replacements {
			if err := replaceTo(replacement.source, replacement.dest, func(data []byte) ([]byte, error) {
				return templateReplacement(data, ctx)
			}); err != nil {
				return err
			}
		}
	}

	return nil
}

