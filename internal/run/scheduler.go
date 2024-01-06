package run

import (
	"context"

	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/workspace"
	"golang.org/x/exp/maps"
)

type scheduler struct {
	runs runLister
}

type runLister interface {
	List(opts ListOptions) []*Run
}

// StartScheduler starts the run scheduler, which is responsible for ensuring
// there is at most one active run on each workspace at a time.
//
// The scheduler attempts to schedule runs upon every run event it receives.
func StartScheduler(ctx context.Context, runs *Service, workspaces *workspace.Service) {
	sub := runs.Broker.Subscribe(ctx)
	s := &scheduler{runs: runs}

	go func() {
		for range sub {
			for _, run := range s.schedule() {
				// Update status from pending to scheduled
				run.updateStatus(Scheduled)
				// Set run as workspace's current run
				workspaces.SetCurrentRun(run.WorkspaceID(), run.ID)
				// Trigger a plan task
				_, _ = runs.plan(run)
			}
		}
	}()
}

// schedule returns runs that are ready to be scheduled.
func (s *scheduler) schedule() []*Run {
	// Retrieve all pending runs
	pending := s.runs.List(ListOptions{
		Status: []Status{Pending},
		// Oldest runs take priority
		Oldest: true,
	})
	if len(pending) == 0 {
		// Nothing more to schedule
		return nil
	}

	// Populate a map of the oldest pending run for each workspace.
	workspacePending := make(map[resource.ID]*Run)
	for _, p := range pending {
		workspaceID := p.WorkspaceID()
		if _, ok := workspacePending[workspaceID]; !ok {
			workspacePending[workspaceID] = p
		}
	}

	// Retrieve all active runs. Active means a run that is neither in a
	// finished state, pending, nor planned.
	active := s.runs.List(ListOptions{
		Status: []Status{
			Scheduled,
			PlanQueued,
			Planning,
			ApplyQueued,
			Applying,
		},
		Oldest: true,
	})
	if len(active) == 0 {
		// Short cut: there aren't any active runs, so we know we can schedule
		// each of the oldest pending runs for each workspace
		return maps.Values(workspacePending)
	}
	// There are active runs, so we need to determine which workspaces they're
	// on before we know whether pending runs are blocked. Therefore we build a
	// set of blocked workspaces
	blocked := make(map[resource.ID]struct{})
	for _, a := range active {
		workspaceID := a.WorkspaceID()
		blocked[workspaceID] = struct{}{}
	}
	// Schedule pending runs that are not on a blocked workspace
	var scheduled []*Run
	for wid, p := range workspacePending {
		if _, ok := blocked[wid]; !ok {
			scheduled = append(scheduled, p)
		}
	}
	return scheduled
}
