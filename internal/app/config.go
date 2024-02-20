package app

import (
	"errors"
	"fmt"
	"os"
	"runtime"

	"github.com/peterbourgon/ff/v4"
	"github.com/peterbourgon/ff/v4/ffhelp"
	"github.com/peterbourgon/ff/v4/ffyaml"
)

type config struct {
	Program  string
	MaxTasks int
}

// set config in order of precedence:
// 1. flags > 2. env vars > 3. config file
func SetConfig(args []string) (config, error) {
	var cfg config

	fs := ff.NewFlagSet("pug")
	fs.StringVar(&cfg.Program, 'p', "program", "terraform", "The default program to use with pug.")
	fs.IntVar(&cfg.MaxTasks, 't', "max-tasks", 2*runtime.NumCPU(), "The maximum number of parallel tasks.")
	_ = fs.String('c', "config", "pug.yaml", "Path to config file.")

	err := ff.Parse(fs, args,
		ff.WithEnvVarPrefix("PUG"),
		ff.WithConfigFileFlag("config"),
		ff.WithConfigFileParser(ffyaml.Parse),
		ff.WithConfigAllowMissingFile(),
	)
	if errors.Is(err, ff.ErrHelp) {
		fmt.Fprintln(os.Stderr, ffhelp.Flags(fs))
		return config{}, nil
	} else if err != nil {
		return config{}, err
	}

	return cfg, nil
}
