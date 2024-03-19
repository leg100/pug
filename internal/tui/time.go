package tui

import (
	"fmt"
	"time"
)

func Ago(now, t time.Time) string {
	diff := now.Sub(t)
	var (
		n      int
		suffix string
	)

	switch {
	case diff < time.Minute:
		n = int(diff.Seconds())
		suffix = "s"
	case diff < time.Hour:
		n = int(diff.Minutes())
		suffix = "m"
	}
	return fmt.Sprintf("%d%s ago", n, suffix)
}
