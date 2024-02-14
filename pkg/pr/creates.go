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

		destPath := []byte(tpl.Destination)
		dest, err := templateReplacement(destPath, ctx)
		if err != nil {
			dest = destPath
		}

		if err := replaceTo(source, string(dest), func(data []byte) ([]byte, error) {
			return templateReplacement(data, ctx)
		}); err != nil {
			return err
		}
	}

	return nil
}
