package pr_test

import (
	"os"
	"path/filepath"
	"testing"

	"gotest.tools/v3/assert"

	"github.com/pluralsh/plural-cli/pkg/pr"
)

func createFile(path, content string) (*os.File, error) {
	f, err := os.Create(path)
	if err != nil {
		return nil, err
	}

	_, err = f.WriteString(content)
	return f, err
}

func createFiles(fileMap map[string]string) (func(), error) {
	files := make([]*os.File, len(fileMap))
	for path, content := range fileMap {
		f, err := createFile(path, content)
		if err != nil {
			return nil, err
		}

		files = append(files, f)
	}

	return func() {
		for _, file := range files {
			file.Close()
		}
	}, nil
}

func TestApply(t *testing.T) {
	dir := t.TempDir()
	cases := []struct {
		name        string
		files       map[string]string
		templateIn  *pr.PrTemplate
		templateOut *pr.PrTemplate
		err         error
	}{
		{
			name:  "should work with single line regex replacements",
			files: map[string]string{
				filepath.Join(dir, "workload.tf"): `
`,
			},
			templateIn: &pr.PrTemplate{
				Context: map[string]interface{}{
					"version": "1.28",
				},
				Spec: pr.PrTemplateSpec{
					Updates: &pr.UpdateSpec{
						Regexes:           nil,
						Files:             nil,
						ReplaceTemplate:   "",
						Yq:                "",
						MatchStrategy:     "",
						RegexReplacements: nil,
					},
				},
			},
			templateOut: &pr.PrTemplate{
				Spec: pr.PrTemplateSpec{
					Updates: &pr.UpdateSpec{
						Regexes:           nil,
						Files:             nil,
						ReplaceTemplate:   "",
						Yq:                "",
						MatchStrategy:     "",
						RegexReplacements: nil,
					},
				},
			},
			err: nil,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			cleanupFunc, err := createFiles(c.files)
			if err != nil {
				t.Fatal(err)
			}

			defer cleanupFunc()
			err = pr.Apply(c.templateIn)

			assert.ErrorIs(t, err, c.err)
			assert.DeepEqual(t, c.templateIn, c.templateOut)
		})
	}
}
