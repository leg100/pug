package main

import (
	"fmt"
	"os"
	"runtime/pprof"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/leg100/pug/internal/app"
	"github.com/leg100/pug/internal/task"
	"github.com/leg100/pug/internal/tui"
)

func main() {
	f, _ := os.Create("cpu.prof")
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()

	if err := start(); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func start() error {
	cfg, err := app.SetConfig(os.Args[1:])
	if err != nil {
		return fmt.Errorf("setting config: %w", err)
	}

	model, err := tui.New(task.NewRunner(cfg.MaxTasks, "tofu"))
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
