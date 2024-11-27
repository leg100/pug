package app

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/hashicorp/terraform/command/cliconfig"
	"github.com/leg100/pug/internal"
	"github.com/leg100/pug/internal/logging"
	"github.com/peterbourgon/ff/v4"
	"github.com/peterbourgon/ff/v4/ffhelp"
	"github.com/peterbourgon/ff/v4/ffyaml"
)

type Config struct {
	Program                 string
	MaxTasks                int
	PluginCache             bool
	Debug                   bool
	DisableReloadAfterApply bool
	Workdir                 internal.Workdir
	DataDir                 string
	Envs                    []string
	Args                    []string
	Terragrunt              bool
	Logging                 logging.Options

	Version bool
}

// set config in order of precedence:
// 1. flags > 2. env vars > 3. config file
func Parse(stderr io.Writer, args []string) (Config, error) {
	var cfg Config

	home, err := os.UserHomeDir()
	if err != nil {
		return Config{}, fmt.Errorf("retrieving user's home directory: %w", err)
	}
	defaultDataDir := filepath.Join(home, ".pug")
	defaultConfigFile := filepath.Join(home, ".pug.yaml")

	fs := ff.NewFlagSet("pug")
	fs.StringVar(&cfg.Program, 'p', "program", "terraform", "The default program to use with pug.")
	workdir := fs.String('w', "workdir", ".", "The working directory containing modules.")
	fs.IntVar(&cfg.MaxTasks, 't', "max-tasks", 2*runtime.NumCPU(), "The maximum number of parallel tasks.")
	fs.StringVar(&cfg.DataDir, 0, "data-dir", defaultDataDir, "Directory in which to store plan files.")
	fs.StringListVar(&cfg.Envs, 'e', "env", "Environment variable to pass to terraform process. Can set more than once.")
	fs.StringListVar(&cfg.Args, 'a', "arg", "CLI arg to pass to terraform process. Can set more than once.")
	fs.BoolVar(&cfg.Debug, 'd', "debug", "Log bubbletea messages to messages.log")
	fs.BoolVar(&cfg.Version, 'v', "version", "Print version.")
	_ = fs.String('c', "config", defaultConfigFile, "Path to config file.")

	fs.BoolVar(&cfg.DisableReloadAfterApply, 0, "disable-reload-after-apply", "Disable automatic reload of state following an apply.")

	{
		usage := fmt.Sprintf("Logging level (valid: %s).", strings.Join(logging.ValidLevels(), ","))
		fs.StringEnumVar(&cfg.Logging.Level, 'l', "log-level", usage, logging.ValidLevels()...)
	}

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
		return Config{}, err
	}

	// If user has specified terragrunt as the program executable then enable
	// terragrunt mode.
	if cfg.Program == "terragrunt" {
		cfg.Terragrunt = true
	}

	// Perform any conversions from the flag parsed primitive types to pug
	// defined types.
	cfg.Workdir, err = internal.NewWorkdir(*workdir)
	if err != nil {
		return Config{}, err
	}

	return cfg, nil
}
