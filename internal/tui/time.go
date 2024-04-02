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
	// If less than 10 seconds, then report number of seconds ago
	case diff < time.Second*10:
		n = int(diff.Seconds())
		suffix = "s"
		// If between 10 seconds and a minute then report number of seconds in 10
		// second blocks. We do this because this func is called on every render,
		// and it can be discombobulating to the user when every row in a table is
		// updating every second as they navigate it...using 10 seconds blocks helps
		// a little.
	case diff < time.Minute:
		n = int(diff.Round(10 * time.Second).Seconds())
		suffix = "s"
	case diff < time.Hour:
		n = int(diff.Minutes())
		suffix = "m"
	default:
		n = int(diff.Hours())
		suffix = "h"
	}
	return fmt.Sprintf("%d%s ago", n, suffix)
}
