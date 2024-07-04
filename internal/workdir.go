package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// Workdir is the working directory for pug, which is the directory in which
// modules are expected to reside.
type Workdir struct {
	path       string
	prettyPath string
}

// NewWorkdir constructs a working directory from the path. The path must exist.
func NewWorkdir(path string) (Workdir, error) {
	// Ensure path exists
	if _, err := os.Stat(path); err != nil {
		return Workdir{}, fmt.Errorf("working directory \"%s\": %w", path, err)
	}

	// Convert path into an absolute path if not already.
	abs, err := filepath.Abs(path)
	if err != nil {
		return Workdir{}, fmt.Errorf("converting working directory \"%s\" to an absolute path: %w", path, err)
	}
	wd := Workdir{path: abs}

	// If the path is in the user's home directory, replace the home directory
	// prefix with a tilde for pretty printing purposes.
	home, err := os.UserHomeDir()
	if err != nil {
		return Workdir{}, fmt.Errorf("retrieving user's home directory: %w", err)
	}
	if strings.HasPrefix(abs, home) {
		wd.prettyPath = "~" + strings.TrimPrefix(abs, home)
	}
	return wd, nil
}

func (wd Workdir) String() string {
	return wd.path
}

func (wd Workdir) PrettyString() string {
	if wd.prettyPath != "" {
		return wd.prettyPath
	}
	return wd.path
}

func NewTestWorkdir(t *testing.T) Workdir {
	wd, err := NewWorkdir(t.TempDir())
	require.NoError(t, err)
	return wd
}
