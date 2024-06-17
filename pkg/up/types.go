package up

type State struct {
	Version string `json:"terraform_version"`
	Values  Values `json:"values"`
}

type Values struct {
	Outputs map[string]Output `json:"outputs"`
}

type Output struct {
	Value     interface{} `json:"value"`
	FieldType interface{} `json:"type"`
	Sensitive bool        `json:"sensitive"`
}

func stacksRole(outs map[string]Output) string {
	def := "dummy"
	if outs == nil {
		return def
	}

	if i, ok := outs["identity"]; ok {
		if id, ok := i.Value.(string); ok {
			return id
		}
	}

	return def
}
