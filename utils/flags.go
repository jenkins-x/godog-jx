package utils

import (
	"os"
	"strings"
)

// DeleteApps should we delete apps after the quickstart has run
func DeleteApps() bool {
	text := os.Getenv("JX_DISABLE_DELETE_APP")
	return strings.ToLower(text) != "true"
}

// DeleteApps should we delete the git repos after the quickstart has run
func DeleteRepos() bool {
	text := os.Getenv("JX_DISABLE_DELETE_REPO")
	return strings.ToLower(text) != "true"
}
