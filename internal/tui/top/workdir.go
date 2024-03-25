package top

import (
	"fmt"
	"os"
	"strings"
)

// contractUserPath detects if a path is within the user's home directory, and
// if so, replaces the home directory component with a tilde.
func contractUserPath(workdir string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("retrieving user's home directory: %w", err)
	}
	if strings.HasPrefix(workdir, home) {
		workdir = "~" + strings.TrimPrefix(workdir, home)
	}
	return workdir, nil
}
