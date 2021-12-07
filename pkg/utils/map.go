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

func Dedupe(l []string) []string {
	res := make([]string, 0)
	seen := make(map[string]bool)
	for _, val := range l {
		if _, ok := seen[val]; ok {
			continue
		}
		res = append(res, val)
		seen[val] = true
	}

	return res
}