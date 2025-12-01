package app

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/leg100/pug/internal"
	"github.com/leg100/pug/internal/logging"
	"github.com/leg100/pug/internal/testutils"
	"github.com/peterbourgon/ff/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig(t *testing.T) {
	// Unset environment variables set on host computer
	testutils.ResetEnv(t)
	t.Setenv("HOME", t.TempDir())

	tests := []struct {
		name string
		file string
		args []string
		envs []string
		want func(t *testing.T, got Config)
	}{
		{
			"defaults",
			"",
			nil,
			nil,
			func(t *testing.T, got Config) {
				pwd, err := os.Getwd()
				require.NoError(t, err)
				wd, err := internal.NewWorkdir(pwd)
				require.NoError(t, err)

				want := Config{
					Program:  "terraform",
					MaxTasks: 2 * runtime.NumCPU(),
					Workdir:  wd,
					DataDir:  filepath.Join(os.Getenv("HOME"), ".pug"),
					Logging: logging.Options{
						Level: "info",
					},
				}
				assert.Equal(t, want, got)
			},
		},
		{
			"config file override default",
			"program: tofu\n",
			nil,
			nil,
			func(t *testing.T, got Config) {
				assert.Equal(t, got.Program, "tofu")
			},
		},
		{
			"config file with max-tasks override default",
			"max-tasks: 3\n",
			nil,
			nil,
			func(t *testing.T, got Config) {
				assert.Equal(t, got.MaxTasks, 3)
			},
		},
		{
			"env var override default",
			"",
			nil,
			[]string{"PUG_PROGRAM=tofu"},
			func(t *testing.T, got Config) {
				assert.Equal(t, got.Program, "tofu")
			},
		},
		{
			"flag override default",
			"",
			[]string{"--program", "tofu"},
			nil,
			func(t *testing.T, got Config) {
				assert.Equal(t, got.Program, "tofu")
			},
		},
		{
			"env var overrides config file",
			"program: tofu\n",
			nil,
			[]string{"PUG_PROGRAM=terragrunt"},
			func(t *testing.T, got Config) {
				assert.Equal(t, got.Program, "terragrunt")
			},
		},
		{
			"flag overrides env var",
			"",
			[]string{"--program", "tofu"},
			[]string{"PUG_PROGRAM=terragrunt"},
			func(t *testing.T, got Config) {
				assert.Equal(t, got.Program, "tofu")
			},
		},
		{
			"flag overrides both env var and config",
			"program: cloudformation\n",
			[]string{"--program", "tofu"},
			[]string{"PUG_PROGRAM=terragrunt"},
			func(t *testing.T, got Config) {
				assert.Equal(t, got.Program, "tofu")
			},
		},
		{
			"enable plugin cache via env var",
			"",
			nil,
			[]string{"TF_PLUGIN_CACHE_DIR=../testdata/plugin_cache"},
			func(t *testing.T, got Config) {
				assert.True(t, got.PluginCache)
			},
		},
		{
			"set terraform process environment variable",
			"",
			[]string{"-e", "TF_LOG=DEBUG"},
			nil,
			func(t *testing.T, got Config) {
				assert.Equal(t, got.Envs, []string{"TF_LOG=DEBUG"})
			},
		},
		{
			"set multiple terraform process environment variables",
			"",
			[]string{"-e", "TF_LOG=DEBUG", "-e", "TF_IGNORE=TRACE", "-e", "TF_PLUGIN_CACHE_DIR=/tmp"},
			nil,
			func(t *testing.T, got Config) {
				assert.Len(t, got.Envs, 3)
				assert.Contains(t, got.Envs, "TF_LOG=DEBUG")
				assert.Contains(t, got.Envs, "TF_IGNORE=TRACE")
				assert.Contains(t, got.Envs, "TF_PLUGIN_CACHE_DIR=/tmp")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// change into a temp dir in case the host computer has a pug.yaml file
			testutils.ChTempDir(t, t.TempDir())

			// set env vars
			for _, ev := range tt.envs {
				name, val, _ := strings.Cut(ev, "=")
				t.Setenv(name, val)
			}

			// set config file
			if tt.file != "" {
				path := filepath.Join(os.Getenv("HOME"), ".pug.yaml")
				err := os.WriteFile(path, []byte(tt.file), 0o644)
				require.NoError(t, err)
			}

			// and pass in flags
			got, err := Parse(io.Discard, tt.args)
			require.NoError(t, err)

			tt.want(t, got)
		})
	}
}

func TestHelpFlag(t *testing.T) {
	for _, flag := range []string{"--help", "-h"} {
		got := new(bytes.Buffer)
		_, err := Parse(got, []string{flag})
		// Help flag should return error
		assert.ErrorIs(t, err, ff.ErrHelp)

		want := "-l, --log-level STRING             Logging level (valid: info,debug,error,warn). (default: info)"
		assert.Contains(t, got.String(), want)
	}
}
