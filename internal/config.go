package internal

import (
	"errors"
	"fmt"
	"os"

	"github.com/peterbourgon/ff/v4"
	"github.com/peterbourgon/ff/v4/ffhelp"
	"github.com/peterbourgon/ff/v4/ffyaml"
)

type config struct {
	Program  string
	MaxTasks int
	LogFile  string
}

// set config in order of precedence:
// 1. flags > 2. env vars > 3. config file
func SetConfig(args []string) (config, error) {
	var cfg config

	fs := ff.NewFlagSet("pug")
	fs.StringVar(&cfg.Program, 'p', "program", "terraform", "The default program to use with pug.")
	fs.StringVar(&cfg.LogFile, 'l', "log-file", "", "Send log output to a file.")
	fs.IntVar(&cfg.MaxTasks, 't', "max-tasks", 5, "The maximum number of parallel tasks.")
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
