package utils

func DeepSet(v map[string]interface{}, path []string, val interface{}) map[string]interface{} {
	key := path[0]
	if len(path) == 1 {
		v[key] = val
		return v
	}

	if next, ok := v[key]; ok {
		switch next.(type) {
		case map[string]interface{}:
			v[key] = DeepSet(next.(map[string]interface{}), path[1:], val)
			return v
		}
	}

	return v
}