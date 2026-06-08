package agents

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/mitchellh/go-homedir"
)

// Confirmer abstracts yes/no prompts used while preparing a local resume.
type Confirmer interface {
	// Confirm asks a yes/no question and returns the selected value.
	Confirm(message string, def bool) (bool, error)
}

// Selector abstracts single-choice prompts.
type Selector interface {
	// Select asks the user to choose one option from the supplied list.
	Select(message string, options []string) (string, error)
}

// DirectoryPrompter abstracts directory selection with validation and completion.
type DirectoryPrompter interface {
	// Directory asks the user for an existing directory path.
	Directory(message, def string) (string, error)
}

// Interaction groups all user prompts needed by agent session restoration.
//
// This interface keeps restore and repository logic independent from the
// concrete terminal UI implementation. The current implementation uses survey,
// but callers should depend on this interface so the prompts can move to a
// Bubble Tea model later without rewriting the restore flow.
type Interaction interface {
	Confirmer
	Selector
	DirectoryPrompter
}

// SurveyInteraction implements Interaction using github.com/AlecAivazis/survey.
type SurveyInteraction struct{}

// NewSurveyInteraction returns the default prompt implementation for CLI use.
func NewSurveyInteraction() *SurveyInteraction {
	return &SurveyInteraction{}
}

func (SurveyInteraction) Confirm(message string, def bool) (bool, error) {
	var confirmed bool
	prompt := &survey.Confirm{Message: message, Default: def}
	if err := survey.AskOne(prompt, &confirmed); err != nil {
		return false, err
	}
	return confirmed, nil
}

func (SurveyInteraction) Select(message string, options []string) (string, error) {
	var selected string
	if err := survey.AskOne(&survey.Select{
		Message: message,
		Options: options,
	}, &selected, survey.WithValidator(survey.Required)); err != nil {
		return "", err
	}
	return selected, nil
}

func (s SurveyInteraction) Directory(message, def string) (string, error) {
	var path string
	if err := survey.AskOne(&survey.Input{
		Message: message,
		Default: def,
		Suggest: s.suggestDirectories,
	}, &path, survey.WithValidator(survey.ComposeValidators(survey.Required, s.directoryValidator))); err != nil {
		return "", err
	}
	expanded, err := homedir.Expand(path)
	if err != nil {
		return "", err
	}
	return filepath.Abs(expanded)
}

func (s SurveyInteraction) suggestDirectories(toComplete string) []string {
	if strings.TrimSpace(toComplete) == "" {
		toComplete = "."
	}
	expanded, err := homedir.Expand(toComplete)
	if err != nil {
		return nil
	}

	searchDir := expanded
	prefix := ""
	if !strings.HasSuffix(expanded, string(filepath.Separator)) {
		searchDir = filepath.Dir(expanded)
		prefix = filepath.Base(expanded)
	}

	entries, err := os.ReadDir(searchDir)
	if err != nil {
		return nil
	}

	suggestions := make([]string, 0)
	for _, entry := range entries {
		if !entry.IsDir() || !strings.HasPrefix(entry.Name(), prefix) {
			continue
		}
		suggestion := filepath.Join(searchDir, entry.Name()) + string(filepath.Separator)
		suggestions = append(suggestions, s.restoreHomePrefix(toComplete, suggestion))
	}
	sort.Strings(suggestions)
	return suggestions
}

func (SurveyInteraction) directoryValidator(value any) error {
	raw, ok := value.(string)
	if !ok {
		return fmt.Errorf("directory path must be a string")
	}
	expanded, err := homedir.Expand(raw)
	if err != nil {
		return err
	}
	info, err := os.Stat(expanded)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", raw)
	}
	return nil
}

func (SurveyInteraction) restoreHomePrefix(input, suggestion string) string {
	if !strings.HasPrefix(input, "~") {
		return suggestion
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return suggestion
	}
	if suggestion == home {
		return "~"
	}
	if strings.HasPrefix(suggestion, home+string(filepath.Separator)) {
		return "~" + strings.TrimPrefix(suggestion, home)
	}
	return suggestion
}
