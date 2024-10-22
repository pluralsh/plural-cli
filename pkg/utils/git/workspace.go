package git

import (
	"strings"
)

func Modified() ([]string, error) {
	res, err := GitRaw("status", "--porcelain")
	if err != nil {
		return nil, err
	}

	result := make([]string, 0)
	for _, line := range strings.Split(res, "\n") {
		cols := strings.Fields(strings.TrimSpace(line))
		if len(cols) > 1 {
			result = append(result, cols[1])
		}
	}

	return result, nil
}

func Status() (string, error) {
	return GitRaw("status", "--porcelain")
}
