package run

type watcher struct {
	svc *Service
}

// Sync synchronises run status with tasks
//func (w *watcher) Sync(event resource.Event[*task.Task]) {
//	switch event.Type {
//	case resource.UpdatedEvent:
//		run, ok := w.svc.runs[event.Payload.Parent]
//		if !ok {
//			return
//		}
//		// TODO: check whether task is for a plan or apply.
//		switch event.Payload.State {
//		case task.Queued:
//			run.Status = PlanQueued
//		case task.Running:
//			run.Status = Planning
//		case task.Canceled:
//			run.Status = Canceled
//		case task.Errored:
//			run.Status = Errored
//		case task.Exited:
//			// TODO: parse plan file and determine whether changes are proposed,
//			// update run accordingly.
//			run.Status = Planned
//		}
//	}
//}
