package internal

import (
	"github.com/btcsuite/btcutil/base58"
	"github.com/google/uuid"
)

type runstate string

const (
	runplanqueued         runstate = "plan_queued"
	runplanning           runstate = "planning"
	runplanned            runstate = "planned"
	runplannedandfinished runstate = "planned_and_finished"
	runapplyqueued        runstate = "apply_queued"
	runapplying           runstate = "applying"
	runapplied            runstate = "applied"
	runerrored            runstate = "errored"
	runcanceled           runstate = "canceled"
)

type run struct {
	id    string
	state runstate
}

func newRun() *run {
	u := uuid.New()
	return &run{
		id:    base58.Encode(u[:]),
		state: runplanqueued,
	}
}
