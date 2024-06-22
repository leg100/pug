package app

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"

	"github.com/hashicorp/terraform/command/cliconfig"
	"github.com/leg100/pug/internal/logging"
	"github.com/peterbourgon/ff/v4"
	"github.com/peterbourgon/ff/v4/ffhelp"
	"github.com/peterbourgon/ff/v4/ffyaml"
)

type config struct {
	Program                 string
	MaxTasks                int
	PluginCache             bool
	FirstPage               string
	Debug                   bool
	DisableReloadAfterApply bool
	WorkDir                 string
	DataDir                 string
	Envs                    []string
	Terragrunt              bool

	loggingOptions logging.Options
	version        bool
}

// set config in order of precedence:
// 1. flags > 2. env vars > 3. config file
func parse(stderr io.Writer, args []string) (config, error) {
	var cfg config

	home, err := os.UserHomeDir()
	if err != nil {
		return config{}, fmt.Errorf("retrieving user's home directory: %w", err)
	}
	defaultDataDir := filepath.Join(home, ".pug")

	fs := ff.NewFlagSet("pug")
	fs.StringVar(&cfg.Program, 'p', "program", "terraform", "The default program to use with pug.")
	fs.StringVar(&cfg.WorkDir, 'w', "workdir", ".", "The working directory containing modules.")
	fs.IntVar(&cfg.MaxTasks, 't', "max-tasks", 2*runtime.NumCPU(), "The maximum number of parallel tasks.")
	fs.StringVar(&cfg.DataDir, 0, "data-dir", defaultDataDir, "Directory in which to store plan files.")
	fs.StringListVar(&cfg.Envs, 'e', "env", "Environment variable to pass to terraform process. Can set more than once.")
	fs.StringEnumVar(&cfg.FirstPage, 'f', "first-page", "The first page to open on startup.", "modules", "workspaces", "runs", "tasks", "logs")
	fs.BoolVar(&cfg.Debug, 'd', "debug", "Log bubbletea messages to messages.log")
	fs.BoolVar(&cfg.version, 'v', "version", "Print version.")
	fs.StringEnumVar(&cfg.loggingOptions.Level, 'l', "log-level", "Logging level.", "info", "debug", "error", "warn")
	_ = fs.String('c', "config", "pug.yaml", "Path to config file.")

	fs.BoolVar(&cfg.DisableReloadAfterApply, 0, "disable-reload-after-apply", "Disable automatic reload of state following an apply.")

	// Plugin cache is enabled not via pug flags but via terraform config
	tfcfg, _ := cliconfig.LoadConfig()
	cfg.PluginCache = (tfcfg.PluginCacheDir != "")

	err = ff.Parse(fs, args,
		ff.WithEnvVarPrefix("PUG"),
		ff.WithConfigFileFlag("config"),
		ff.WithConfigFileParser(ffyaml.Parse),
		ff.WithConfigAllowMissingFile(),
	)
	if err != nil {
		// ff.Parse returns an error if there is an error or if -h/--help is
		// passed; in either case print flag usage in addition to error message.
		fmt.Fprintln(stderr, ffhelp.Flags(fs))
		return config{}, err
	}

	// If user has specified terragrunt as the program executable then enable
	// terragrunt mode.
	if cfg.Program == "terragrunt" {
		cfg.Terragrunt = true
	}

	return cfg, nil
}
