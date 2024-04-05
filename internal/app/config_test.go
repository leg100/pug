package app

import (
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig(t *testing.T) {
	// Unset environment variables set on host computer
	t.Setenv("PUG_DEBUG", "")
	t.Setenv("PUG_FIRST_PAGE", "")
	t.Setenv("PUG_LOG_LEVEL", "")

	tests := []struct {
		name string
		file string
		args []string
		envs []string
		want func(t *testing.T, got config)
	}{
		{
			"defaults",
			"",
			nil,
			nil,
			func(t *testing.T, got config) {
				want := config{
					Program:   "terraform",
					MaxTasks:  2 * runtime.NumCPU(),
					FirstPage: "modules",
					LogLevel:  "info",
				}
				assert.Equal(t, want, got)
			},
		},
		{
			"config file override default",
			"program: tofu\n",
			nil,
			nil,
			func(t *testing.T, got config) {
				assert.Equal(t, got.Program, "tofu")
			},
		},
		{
			"config file with max-tasks override default",
			"max-tasks: 3\n",
			nil,
			nil,
			func(t *testing.T, got config) {
				assert.Equal(t, got.MaxTasks, 3)
			},
		},
		{
			"env var override default",
			"",
			nil,
			[]string{"PUG_PROGRAM=tofu"},
			func(t *testing.T, got config) {
				assert.Equal(t, got.Program, "tofu")
			},
		},
		{
			"flag override default",
			"",
			[]string{"--program", "tofu"},
			nil,
			func(t *testing.T, got config) {
				assert.Equal(t, got.Program, "tofu")
			},
		},
		{
			"env var overrides config file",
			"program: tofu\n",
			nil,
			[]string{"PUG_PROGRAM=terragrunt"},
			func(t *testing.T, got config) {
				assert.Equal(t, got.Program, "terragrunt")
			},
		},
		{
			"flag overrides env var",
			"",
			[]string{"--program", "tofu"},
			[]string{"PUG_PROGRAM=terragrunt"},
			func(t *testing.T, got config) {
				assert.Equal(t, got.Program, "tofu")
			},
		},
		{
			"flag overrides both env var and config",
			"program: cloudformation\n",
			[]string{"--program", "tofu"},
			[]string{"PUG_PROGRAM=terragrunt"},
			func(t *testing.T, got config) {
				assert.Equal(t, got.Program, "tofu")
			},
		},
		{
			"enable plugin cache via env var",
			"",
			nil,
			[]string{"TF_PLUGIN_CACHE_DIR=../testdata/plugin_cache"},
			func(t *testing.T, got config) {
				assert.True(t, got.PluginCache)
			},
		},
		{
			"set first page via environment variable",
			"",
			nil,
			[]string{"PUG_FIRST_PAGE=runs"},
			func(t *testing.T, got config) {
				assert.Equal(t, got.FirstPage, "runs")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// change into a temp dir in case the host computer has a pug.yaml file
			err := os.Chdir(t.TempDir())
			require.NoError(t, err)

			// set env vars
			for _, ev := range tt.envs {
				name, val, _ := strings.Cut(ev, "=")
				t.Setenv(name, val)
			}

			// set config file
			if tt.file != "" {
				err := os.WriteFile("pug.yaml", []byte(tt.file), 0o400)
				require.NoError(t, err)
			}

			// and pass in flags
			got, err := parse(tt.args)
			require.NoError(t, err)

			tt.want(t, got)
		})
	}
}
