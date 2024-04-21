package pr

import (
	"os"
	"path/filepath"

	"github.com/osteele/liquid"
)

var (
	liquidEngine = liquid.NewEngine()
)

func templateReplacement(data []byte, ctx map[string]interface{}) ([]byte, error) {
	bindings := map[string]interface{}{
		"context": ctx,
	}
	return liquidEngine.ParseAndRender(data, bindings)
}

func replaceTo(from, to string, rep func(data []byte) ([]byte, error)) error {
	info, err := os.Stat(from)
	if err != nil {
		return err
	}

	data, err := os.ReadFile(from)
	if err != nil {
		return err
	}

	resData, err := rep(data)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(to), 0755); err != nil {
		return err
	}

	return os.WriteFile(to, resData, info.Mode())
}

func replaceInPlace(path string, rep func(data []byte) ([]byte, error)) error {
	return replaceTo(path, path, rep)
}
