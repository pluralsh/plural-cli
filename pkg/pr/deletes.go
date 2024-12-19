package pr

import (
	"os"
)

func applyDeletes(deletes *DeleteSpec, ctx map[string]interface{}) error {
	if deletes == nil {
		return nil
	}

	for _, f := range deletes.Files {
		dest, err := templateReplacement([]byte(f), ctx)
		if err != nil {
			dest = []byte(f)
		}

		if err := removeMatches(string(dest)); err != nil {
			return err
		}
	}

	for _, f := range deletes.Folders {
		dest, err := templateReplacement([]byte(f), ctx)
		if err != nil {
			dest = []byte(f)
		}

		if err := os.RemoveAll(string(dest)); err != nil {
			return err
		}
	}

	return nil
}
