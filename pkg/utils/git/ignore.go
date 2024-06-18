package git

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
)

// AppendGitIgnore creates or appends to an existing '.gitignore' file.
func AppendGitIgnore(dir string, entries []string) (err error) {
	filePath := filepath.Join(dir, ".gitignore")

	if err = os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	f, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	defer func() {
		err = f.Close()
	}()

	scanner := bufio.NewScanner(f)
	set := make(map[string]struct{})
	for scanner.Scan() {
		set[scanner.Text()] = struct{}{}
	}

	for _, entry := range entries {
		if _, exists := set[entry]; exists {
			continue
		}

		if _, err = f.WriteString(fmt.Sprintf("%s\n", entry)); err != nil {
			return err
		}
	}

	return err
}
