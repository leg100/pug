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
	cpus := runtime.NumCPU()

	tests := []struct {
		name string
		file string
		args []string
		envs []string
		want config
	}{
		{
			"defaults",
			"",
			nil,
			nil,
			config{Program: "terraform", MaxTasks: 2 * cpus},
		},
		{
			"config file override default",
			"program: tofu\n",
			nil,
			nil,
			config{Program: "tofu", MaxTasks: 2 * cpus},
		},
		{
			"config file with max-tasks override default",
			"max-tasks: 3\n",
			nil,
			nil,
			config{Program: "terraform", MaxTasks: 3},
		},
		{
			"env var override default",
			"",
			nil,
			[]string{"PUG_PROGRAM=tofu"},
			config{Program: "tofu", MaxTasks: 2 * cpus},
		},
		{
			"flag override default",
			"",
			[]string{"--program", "tofu"},
			nil,
			config{Program: "tofu", MaxTasks: 2 * cpus},
		},
		{
			"env var overrides config file",
			"program: tofu\n",
			nil,
			[]string{"PUG_PROGRAM=terragrunt"},
			config{Program: "terragrunt", MaxTasks: 2 * cpus},
		},
		{
			"flag overrides env var",
			"",
			[]string{"--program", "tofu"},
			[]string{"PUG_PROGRAM=terragrunt"},
			config{Program: "tofu", MaxTasks: 2 * cpus},
		},
		{
			"flag overrides both env var and config",
			"program: cloudformation\n",
			[]string{"--program", "tofu"},
			[]string{"PUG_PROGRAM=terragrunt"},
			config{Program: "tofu", MaxTasks: 2 * cpus},
		},
		{
			"enable plugin cache via terraform config",
			"",
			nil,
			[]string{"TF_CLI_CONFIG_FILE=../testdata/_terraformrc_with_plugin_cache"},
			config{Program: "terraform", MaxTasks: 2 * cpus, PluginCache: true},
		},
		{
			"enable plugin cache via env var",
			"",
			nil,
			[]string{"TF_PLUGIN_CACHE_DIR=../testdata/plugin_cache"},
			config{Program: "terraform", MaxTasks: 2 * cpus, PluginCache: true},
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
				os.WriteFile("pug.yaml", []byte(tt.file), 0o400)
			}

			// and pass in flags
			got, err := parse(tt.args)
			require.NoError(t, err)

			assert.Equal(t, tt.want, got)
		})
	}
}
