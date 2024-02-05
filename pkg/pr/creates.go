package pr

import (
	"path/filepath"
)

func applyCreates(creates *CreateSpec, ctx map[string]interface{}) error {
	if creates == nil {
		return nil
	}

	for _, tpl := range creates.Templates {
		source := tpl.Source
		if tpl.External {
			source = filepath.Join(creates.ExternalDir, source)
		}

		if err := replaceTo(source, tpl.Destination, func(data []byte) ([]byte, error) {
			return templateReplacement(data, ctx)
		}); err != nil {
			return err
		}
	}

	return nil
}
