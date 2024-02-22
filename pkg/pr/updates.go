package pr

import (
	"io/fs"
	"path/filepath"
	"regexp"
)

func applyUpdates(updates *UpdateSpec, ctx map[string]interface{}) error {
	if updates == nil {
		return nil
	}

	replacement, err := templateReplacement([]byte(updates.ReplaceTemplate), ctx)
	if err != nil {
		return err
	}

	return filepath.Walk(".", func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		ok, err := filenameMatches(path, updates.Files)
		if err != nil {
			return err
		}

		if ok {
			return updateFile(path, updates, replacement)
		}

		return nil
	})
}

func updateFile(path string, updates *UpdateSpec, replacement []byte) error {
	switch updates.MatchStrategy {
	case "any":
		return anyUpdateFile(path, updates, replacement)
	case "all":
		return allUpdateFile(path, updates)
	case "recursive":
		return recursiveUpdateFile(path, updates, replacement)
	default:
		return nil
	}
}

func anyUpdateFile(path string, updates *UpdateSpec, replacement []byte) error {
	return replaceInPlace(path, func(data []byte) ([]byte, error) {
		for _, reg := range updates.Regexes {
			r, err := regexp.Compile(reg)
			if err != nil {
				return data, err
			}
			data = r.ReplaceAll(data, replacement)
		}
		return data, nil
	})
}

func allUpdateFile(path string, updates *UpdateSpec) error {
	return nil
}

func recursiveUpdateFile(path string, updates *UpdateSpec, replacement []byte) error {
	return replaceInPlace(path, func(data []byte) ([]byte, error) {
		return recursiveReplace(data, updates.Regexes, replacement)
	})
}

func recursiveReplace(data []byte, regexes []string, replacement []byte) ([]byte, error) {
	if len(regexes) == 0 {
		return []byte(replacement), nil
	}

	r, err := regexp.Compile(regexes[0])
	if err != nil {
		return data, err
	}

	res := r.ReplaceAllFunc(data, func(d []byte) []byte {
		res, err := recursiveReplace(d, regexes[1:], replacement)
		if err != nil {
			panic(err)
		}
		return res
	})

	return res, nil
}

func filenameMatches(path string, files []string) (bool, error) {
	for _, f := range files {
		r, err := regexp.Compile(f)
		if err != nil {
			return false, err
		}

		if r.MatchString(path) {
			return true, nil
		}
	}

	return false, nil
}
