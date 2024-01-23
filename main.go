package main

import (
	"fmt"
	"io"
	"log/slog"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/leg100/pug/internal"
)

func main() {
	if err := start(); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func start() error {
	cfg, err := internal.SetConfig(os.Args[1:])
	if err != nil {
		return fmt.Errorf("setting config: %w", err)
	}
	if cfg.LogFile != "" {
		f, err := os.OpenFile(cfg.LogFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o755)
		if err != nil {
			return fmt.Errorf("opening file for logging: %w", err)
		}
		logger := slog.New(slog.NewTextHandler(f, nil))
		slog.SetDefault(logger)
	} else {
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		slog.SetDefault(logger)
	}

	model, err := internal.NewMainModel(internal.NewRunner(cfg.MaxTasks))
	if err != nil {
		return err
	}

	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),       // use the full size of the terminal in its "alternate screen buffer"
		tea.WithMouseCellMotion(), // turn on mouse support so we can track the mouse wheel
	)
	if _, err := p.Run(); err != nil {
		return err
	}
	return nil
}
