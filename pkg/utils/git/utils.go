package git

import (
	"regexp"
	"strings"
)

func RepoName(url string) string {
	reg, _ := regexp.Compile(".*/")
	base := reg.ReplaceAllString(url, "")
	return strings.TrimSuffix(base, ".git")
}

func IsSha(str string) bool {
	matches, _ := regexp.MatchString("[a-f0-9]{40}", str)
	return matches
}