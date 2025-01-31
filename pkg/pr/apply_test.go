package pr_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pluralsh/polly/algorithms"
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

func createFiles(dir string, fileMap map[string]string) (func(), error) {
	files := make([]*os.File, len(fileMap))
	for path, content := range fileMap {
		f, err := createFile(filepath.Join(dir, path), content)
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

func readFiles(dir string, paths []string) (map[string]string, error) {
	files := make(map[string]string, len(paths))
	for _, path := range paths {
		content, err := os.ReadFile(filepath.Join(dir, path))
		if err != nil {
			return nil, err
		}

		files[path] = string(content)
	}

	return files, nil
}

func countFiles(dir string) (int, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return 0, err
	}

	return len(files), nil
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
	cases := []struct {
		name          string
		files         map[string]string
		template      *pr.PrTemplate
		expectedFiles map[string]string
		expectedDir   string
		expectedErr   error
	}{
		{
			name: "should work with single line regex replacements",
			files: map[string]string{
				"workload.tf": `
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
								File:        "workload.tf",
								Templated:   false,
							},
						},
					},
				},
			},
			expectedFiles: map[string]string{
				"workload.tf": `
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
				"base.yaml": baseYAMLIn,
			},
			template: &pr.PrTemplate{
				Context: map[string]interface{}{
					"version": "1.28",
				},
				Spec: pr.PrTemplateSpec{
					Updates: &pr.UpdateSpec{
						YamlOverlays: []pr.YamlOverlay{
							{
								File:      "base.yaml",
								Yaml:      overlayYAML,
								ListMerge: pr.ListMergeOverwrite,
								Templated: true,
							},
						},
					},
				},
			},
			expectedFiles: map[string]string{
				"base.yaml": baseYAMLTemplated,
			},
			expectedErr: nil,
		},
		{
			name: "should not template and overlay with overwrite yaml file",
			files: map[string]string{
				"base.yaml": baseYAMLIn,
			},
			template: &pr.PrTemplate{
				Context: map[string]interface{}{
					"version": "1.28",
				},
				Spec: pr.PrTemplateSpec{
					Updates: &pr.UpdateSpec{
						YamlOverlays: []pr.YamlOverlay{
							{
								File:      "base.yaml",
								Yaml:      overlayYAML,
								ListMerge: pr.ListMergeOverwrite,
								Templated: false,
							},
						},
					},
				},
			},
			expectedFiles: map[string]string{
				"base.yaml": baseYAMLNonTemplated,
			},
			expectedErr: nil,
		},
		{
			name: "should template and overlay with append yaml file",
			files: map[string]string{
				"base.yaml": baseYAMLIn,
			},
			template: &pr.PrTemplate{
				Context: map[string]interface{}{
					"version": "1.28",
				},
				Spec: pr.PrTemplateSpec{
					Updates: &pr.UpdateSpec{
						YamlOverlays: []pr.YamlOverlay{
							{
								File:      "base.yaml",
								Yaml:      overlayYAML,
								ListMerge: pr.ListMergeAppend,
								Templated: true,
							},
						},
					},
				},
			},
			expectedFiles: map[string]string{
				"base.yaml": baseYAMLAppend,
			},
			expectedErr: nil,
		},
		{
			name: "should not template and overlay with append yaml file",
			files: map[string]string{
				"base.yaml": baseYAMLIn,
			},
			template: &pr.PrTemplate{
				Context: map[string]interface{}{
					"version": "1.28",
				},
				Spec: pr.PrTemplateSpec{
					Updates: &pr.UpdateSpec{
						YamlOverlays: []pr.YamlOverlay{
							{
								File:      "base.yaml",
								Yaml:      overlayYAML,
								ListMerge: pr.ListMergeAppend,
								Templated: false,
							},
						},
					},
				},
			},
			expectedFiles: map[string]string{
				"base.yaml": baseYAMLAppendNonTemplated,
			},
			expectedErr: nil,
		},
		{
			name: "should not skip file when condition field is missing",
			files: map[string]string{
				"base.yaml": baseYAMLIn,
			},
			template: &pr.PrTemplate{
				Context: map[string]interface{}{
					"version": "1.28",
				},
				Spec: pr.PrTemplateSpec{
					Creates: &pr.CreateSpec{
						Templates: []*pr.CreateTemplate{
							{
								Source:      "base.yaml",
								Destination: "base.yaml",
							},
						},
					},
				},
			},
			expectedFiles: map[string]string{
				"base.yaml": baseYAMLIn,
			},
			expectedDir: t.TempDir(),
			expectedErr: nil,
		},
		{
			name: "should not skip file when condition field is empty",
			files: map[string]string{
				"base.yaml": baseYAMLIn,
			},
			template: &pr.PrTemplate{
				Context: map[string]interface{}{
					"version": "1.28",
				},
				Spec: pr.PrTemplateSpec{
					Creates: &pr.CreateSpec{
						Templates: []*pr.CreateTemplate{
							{
								Source:      "base.yaml",
								Destination: "base.yaml",
								Condition:   "",
							},
						},
					},
				},
			},
			expectedFiles: map[string]string{
				"base.yaml": baseYAMLIn,
			},
			expectedDir: t.TempDir(),
			expectedErr: nil,
		},
		{
			name: "should not skip file when condition field evaluates to true",
			files: map[string]string{
				"base.yaml": baseYAMLIn,
			},
			template: &pr.PrTemplate{
				Context: map[string]interface{}{
					"version": "1.28",
				},
				Spec: pr.PrTemplateSpec{
					Creates: &pr.CreateSpec{
						Templates: []*pr.CreateTemplate{
							{
								Source:      "base.yaml",
								Destination: "base.yaml",
								Condition:   "context.version == `1.28`",
							},
						},
					},
				},
			},
			expectedFiles: map[string]string{
				"base.yaml": baseYAMLIn,
			},
			expectedDir: t.TempDir(),
			expectedErr: nil,
		},
		{
			name: "should skip file when condition field evaluates to false",
			files: map[string]string{
				"base.yaml": baseYAMLIn,
			},
			template: &pr.PrTemplate{
				Context: map[string]interface{}{
					"version": "1.28",
				},
				Spec: pr.PrTemplateSpec{
					Creates: &pr.CreateSpec{
						Templates: []*pr.CreateTemplate{
							{
								Source:      "base.yaml",
								Destination: "base.yaml",
								Condition:   "context.version != `1.28`",
							},
						},
					},
				},
			},
			expectedFiles: map[string]string{},
			expectedDir: t.TempDir(),
			expectedErr: nil,
		},
	}

	t.Parallel()
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			dir := t.TempDir()
			outdir := lo.Ternary(len(c.expectedDir) > 0, c.expectedDir, dir)

			cleanupFunc, err := createFiles(dir, c.files)
			assert.NilError(t, err)
			defer cleanupFunc()

			err = pr.Apply(transform(c.template, dir, outdir))
			assert.ErrorIs(t, err, c.expectedErr)

			files, err := readFiles(dir, lo.Keys(c.expectedFiles))
			assert.NilError(t, err)

			fileCount, err := countFiles(outdir)
			assert.NilError(t, err)

			assert.Equal(t, fileCount, len(c.expectedFiles))

			for file, content := range files {
				expectedContent, exists := c.expectedFiles[file]
				assert.Check(t, exists)
				assert.Equal(t, content, expectedContent)
			}
		})
	}
}

// transform updates file paths in PrTemplate since we want
// to be able to dynamically provide the input/output dirs.
// If tests would operate on same directories, it would cause conflicts.
func transform(template *pr.PrTemplate, dir, outdir string) *pr.PrTemplate {
	if template.Spec.Creates != nil {
		template.Spec.Creates.Templates = algorithms.Map(template.Spec.Creates.Templates, func(cTemplate *pr.CreateTemplate) *pr.CreateTemplate {
			cTemplate.Source = filepath.Join(dir, cTemplate.Source)
			cTemplate.Destination = filepath.Join(outdir, cTemplate.Destination)

			return cTemplate
		})
	}

	if template.Spec.Updates != nil {
		template.Spec.Updates.Files = algorithms.Map(template.Spec.Updates.Files, func(f string) string {
			return filepath.Join(dir, f)
		})

		template.Spec.Updates.RegexReplacements = algorithms.Map(template.Spec.Updates.RegexReplacements, func(r pr.RegexReplacement) pr.RegexReplacement {
			r.File = filepath.Join(dir, r.File)
			return r
		})

		template.Spec.Updates.YamlOverlays = algorithms.Map(template.Spec.Updates.YamlOverlays, func(r pr.YamlOverlay) pr.YamlOverlay {
			r.File = filepath.Join(dir, r.File)
			return r
		})
	}

	if template.Spec.Deletes != nil {
		template.Spec.Deletes.Files = algorithms.Map(template.Spec.Deletes.Files, func(f string) string {
			return filepath.Join(dir, f)
		})
	}

	return template
}
