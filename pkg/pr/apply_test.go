package pr_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/samber/lo"
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

func readFiles(paths []string) (map[string]string, error) {
	files := make(map[string]string, len(paths))
	for _, path := range paths {
		content, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}

		files[path] = string(content)
	}

	return files, nil
}

// Notes:
//   - YAML encoder adds a new line at the end!
//   - YAML encoder can reorder fields compared to the overlay YAML.
//     Output YAML field order is stable though.
const (
	baseYAMLIn = `include:
  - directory: foo/foo1
  - directory: foo/foo2
stringtest: old`

	overlayYAML = `include:
  - directory: foo/foo1
    extra: true
    version: '{{ context.version }}'
    stuff:
      stuff1: true
      stuff2: true
  - directory: something/else
stringtest: new
nulltest: ~`

	baseYAMLTemplated = `include:
  - directory: foo/foo1
    extra: true
    stuff:
      stuff1: true
      stuff2: true
    version: "1.28"
  - directory: something/else
nulltest: null
stringtest: new
`

	baseYAMLNonTemplated = `include:
  - directory: foo/foo1
    extra: true
    stuff:
      stuff1: true
      stuff2: true
    version: '{{ context.version }}'
  - directory: something/else
nulltest: null
stringtest: new
`

	baseYAMLAppend = `include:
  - directory: foo/foo1
  - directory: foo/foo2
  - directory: foo/foo1
    extra: true
    stuff:
      stuff1: true
      stuff2: true
    version: "1.28"
  - directory: something/else
nulltest: null
stringtest: new
`

	baseYAMLAppendNonTemplated = `include:
  - directory: foo/foo1
  - directory: foo/foo2
  - directory: foo/foo1
    extra: true
    stuff:
      stuff1: true
      stuff2: true
    version: '{{ context.version }}'
  - directory: something/else
nulltest: null
stringtest: new
`
)

func TestApply(t *testing.T) {
	dir := t.TempDir()
	cases := []struct {
		name          string
		files         map[string]string
		template      *pr.PrTemplate
		expectedFiles map[string]string
		expectedErr   error
	}{
		{
			name: "should work with single line regex replacements",
			files: map[string]string{
				filepath.Join(dir, "workload.tf"): `
					module "staging" {
					  source       = "./eks"
					  cluster_name = "boot-staging"
					  vpc_name     = "plural-stage"
					  kubernetes_version = "1.22"
					  create_db    = false
					  providers = {
						helm = helm.staging
					  }
					}`,
			},
			template: &pr.PrTemplate{
				Context: map[string]interface{}{
					"version": "1.28",
				},
				Spec: pr.PrTemplateSpec{
					Updates: &pr.UpdateSpec{
						RegexReplacements: []pr.RegexReplacement{
							{
								Regex:       "kubernetes_version = \"1.[0-9]+\"",
								Replacement: "kubernetes_version = \"{{ context.version }}\"",
								File:        filepath.Join(dir, "workload.tf"),
								Templated:   false,
							},
						},
					},
				},
			},
			expectedFiles: map[string]string{
				filepath.Join(dir, "workload.tf"): `
					module "staging" {
					  source       = "./eks"
					  cluster_name = "boot-staging"
					  vpc_name     = "plural-stage"
					  kubernetes_version = "1.28"
					  create_db    = false
					  providers = {
						helm = helm.staging
					  }
					}`,
			},
			expectedErr: nil,
		},
		{
			name: "should template and overlay with overwrite yaml file",
			files: map[string]string{
				filepath.Join(dir, "base.yaml"): baseYAMLIn,
			},
			template: &pr.PrTemplate{
				Context: map[string]interface{}{
					"version": "1.28",
				},
				Spec: pr.PrTemplateSpec{
					Updates: &pr.UpdateSpec{
						YamlOverlays: []pr.YamlOverlay{
							{
								File:      filepath.Join(dir, "base.yaml"),
								Yaml:      overlayYAML,
								ListMerge: pr.ListMergeOverwrite,
								Templated: true,
							},
						},
					},
				},
			},
			expectedFiles: map[string]string{
				filepath.Join(dir, "base.yaml"): baseYAMLTemplated,
			},
			expectedErr: nil,
		},
		{
			name: "should not template and overlay with overwrite yaml file",
			files: map[string]string{
				filepath.Join(dir, "base.yaml"): baseYAMLIn,
			},
			template: &pr.PrTemplate{
				Context: map[string]interface{}{
					"version": "1.28",
				},
				Spec: pr.PrTemplateSpec{
					Updates: &pr.UpdateSpec{
						YamlOverlays: []pr.YamlOverlay{
							{
								File:      filepath.Join(dir, "base.yaml"),
								Yaml:      overlayYAML,
								ListMerge: pr.ListMergeOverwrite,
								Templated: false,
							},
						},
					},
				},
			},
			expectedFiles: map[string]string{
				filepath.Join(dir, "base.yaml"): baseYAMLNonTemplated,
			},
			expectedErr: nil,
		},
		{
			name: "should template and overlay with append yaml file",
			files: map[string]string{
				filepath.Join(dir, "base.yaml"): baseYAMLIn,
			},
			template: &pr.PrTemplate{
				Context: map[string]interface{}{
					"version": "1.28",
				},
				Spec: pr.PrTemplateSpec{
					Updates: &pr.UpdateSpec{
						YamlOverlays: []pr.YamlOverlay{
							{
								File:      filepath.Join(dir, "base.yaml"),
								Yaml:      overlayYAML,
								ListMerge: pr.ListMergeAppend,
								Templated: true,
							},
						},
					},
				},
			},
			expectedFiles: map[string]string{
				filepath.Join(dir, "base.yaml"): baseYAMLAppend,
			},
			expectedErr: nil,
		},
		{
			name: "should not template and overlay with append yaml file",
			files: map[string]string{
				filepath.Join(dir, "base.yaml"): baseYAMLIn,
			},
			template: &pr.PrTemplate{
				Context: map[string]interface{}{
					"version": "1.28",
				},
				Spec: pr.PrTemplateSpec{
					Updates: &pr.UpdateSpec{
						YamlOverlays: []pr.YamlOverlay{
							{
								File:      filepath.Join(dir, "base.yaml"),
								Yaml:      overlayYAML,
								ListMerge: pr.ListMergeAppend,
								Templated: false,
							},
						},
					},
				},
			},
			expectedFiles: map[string]string{
				filepath.Join(dir, "base.yaml"): baseYAMLAppendNonTemplated,
			},
			expectedErr: nil,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			cleanupFunc, err := createFiles(c.files)
			assert.NilError(t, err)
			defer cleanupFunc()

			err = pr.Apply(c.template)
			assert.ErrorIs(t, err, c.expectedErr)

			files, err := readFiles(lo.Keys(c.expectedFiles))
			assert.NilError(t, err)

			for file, content := range files {
				expectedContent, exists := c.expectedFiles[file]
				assert.Check(t, exists)
				assert.Equal(t, content, expectedContent)
			}
		})
	}
}
