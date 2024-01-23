package pr

import (
	"os"

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

func replaceInPlace(path string, rep func(data []byte) ([]byte, error)) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	resData, err := rep(data)
	if err != nil {
		return err
	}
	return os.WriteFile(path, resData, info.Mode())
}
