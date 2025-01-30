package pr

import (
	"bytes"
	"io/fs"
	"path/filepath"
	"regexp"
	"strings"

	"dario.cat/mergo"
	"github.com/samber/lo"
	"gopkg.in/yaml.v3"
)

func applyUpdates(updates *UpdateSpec, ctx map[string]interface{}) error {
	if updates == nil {
		return nil
	}

	if err := processRegexReplacements(updates.RegexReplacements, ctx); err != nil {
		return err
	}

	if err := processYamlOverlays(updates.YamlOverlays, ctx); err != nil {
		return err
	}

	replacement, err := templateReplacement([]byte(updates.ReplaceTemplate), ctx)
	if err != nil {
		return err
	}

	files := lo.Map(updates.Files, func(name string, ind int) string {
		res, err := templateReplacement([]byte(name), ctx)
		if err != nil {
			return name
		}
		return string(res)
	})

	return filepath.Walk(".", func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		ok, err := filenameMatches(path, files)
		if err != nil {
			return err
		}

		if ok {
			return updateFile(path, updates, replacement)
		}

		return nil
	})
}

func processRegexReplacements(replacements []RegexReplacement, ctx map[string]interface{}) error {
	if len(replacements) == 0 {
		return nil
	}

	for _, replacement := range replacements {
		replaceWith, err := templateReplacement([]byte(replacement.Replacement), ctx)
		if err != nil {
			return err
		}

		replaceFunc := func(data []byte) ([]byte, error) {
			rx, err := templateReplacement([]byte(replacement.Regex), ctx)
			if err != nil {
				rx = []byte(replacement.Regex)
			}
			return replaceMultiline(rx, replaceWith, data)
		}

		dest, err := templateReplacement([]byte(replacement.File), ctx)
		if err != nil {
			dest = []byte(replacement.File)
		}

		if err := replaceInPlace(string(dest), replaceFunc); err != nil {
			return err
		}
	}

	return nil
}

// replaceMultiline applies regex-based replacements on a multiline string
func replaceMultiline(pattern, replacement, text []byte) ([]byte, error) {
	// Replace all newlines with a unique placeholder
	placeholder := "__NL__"
	flattenedText := strings.ReplaceAll(string(text), "\n", placeholder)

	// Apply the regex replacement on the flattened text
	re, err := regexp.Compile(string(pattern))
	if err != nil {
		return nil, err
	}
	flattenedText = re.ReplaceAllString(flattenedText, string(replacement))

	// Revert the placeholder back to newlines
	finalText := strings.ReplaceAll(flattenedText, placeholder, "\n")
	return []byte(finalText), nil
}

func processYamlOverlays(overlays []YamlOverlay, ctx map[string]interface{}) error {
	if len(overlays) == 0 {
		return nil
	}

	for _, overlay := range overlays {
		var err error
		var overlayYaml = []byte(overlay.Yaml)

		if overlay.Templated {
			overlayYaml, err = templateReplacement([]byte(overlay.Yaml), ctx)
			if err != nil {
				return err
			}
		}

		mergeFunc := func(data []byte) ([]byte, error) {
			return mergeYaml(data, overlayYaml, overlay.ListMerge)
		}

		fileName := overlay.File
		templated, err := templateReplacement([]byte(fileName), ctx)
		if err == nil {
			fileName = string(templated)
		}

		if err = replaceInPlace(fileName, mergeFunc); err != nil {
			return err
		}
	}

	return nil
}

func mergeYaml(base, overlay []byte, merge ListMerge) ([]byte, error) {
	baseMap := make(map[string]interface{})
	overlayMap := make(map[string]interface{})

	if err := yaml.Unmarshal(base, &baseMap); err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(overlay, &overlayMap); err != nil {
		return nil, err
	}

	options := []func(*mergo.Config){mergo.WithOverride, mergo.WithSliceDeepCopy}
	if merge == ListMergeAppend {
		options = append(options, mergo.WithAppendSlice)
	}

	if err := mergo.Merge(
		&baseMap,
		overlayMap,
		options...,
	); err != nil {
		return nil, err
	}

	var b bytes.Buffer
	encoder := yaml.NewEncoder(&b)
	encoder.SetIndent(2)
	if err := encoder.Encode(baseMap); err != nil {
		return nil, err
	}

	defer encoder.Close()
	return b.Bytes(), nil
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
