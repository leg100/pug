package main

import (
	"fmt"
	"log"
	"os"
	"runtime/pprof"

	"github.com/leg100/pug/internal/app"
	"github.com/leg100/pug/internal/tui/top"
	"github.com/leg100/pug/internal/version"
)

func main() {
	f, err := os.Create("cpu.prof")
	if err != nil {
		log.Fatal(err)
	}
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()

	if err := run(); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func run() error {
	// Parse configuration from env vars and flags.
	cfg, err := app.Parse(os.Stderr, os.Args[1:])
	if err != nil {
		return err
	}
	if cfg.Version {
		fmt.Fprintln(os.Stdout, "pug", version.Version)
		return nil
	}
	// Start TUI and block til user exits.
	return top.Start(cfg)
}
