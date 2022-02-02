package output

import (
	"encoding/json"
	"os/exec"
)

type TerraformOutputItem struct {
	Value interface{}
	Type  interface{}
}

func TerraformOutput(path string) (out map[string]interface{}, err error) {
	cmd := exec.Command("terraform", "output", "-json")
	cmd.Dir = path
	res, err := cmd.Output()
	if err != nil {
		return
	}

	outType := make(map[string]TerraformOutputItem)
	err = json.Unmarshal(res, &outType)
	if err != nil {
		return
	}

	out = make(map[string]interface{})
	for key, val := range outType {
		out[key] = val.Value
	}
	return
}
