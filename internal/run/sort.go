package run

func ByUpdatedDesc(i, j *Run) int {
	if i.Updated.Before(j.Updated) {
		return 1
	}
	return -1
}

// ByStatus sorts runs according to the following order:
//
// * Planning, Applying (most recent first)
// * Planned (most recent first)
// * PlanQueued,ApplyQueued (oldest first)
// * Scheduled (oldest first)
// * Pending (oldest first)
// * Applied,Errored,Canceled,Finished,PlannedAndFinished,Discarded (most recent
// first)
func ByStatus(i, j *Run) int {
	switch i.Status {
	case Pending:
		switch j.Status {
		case Pending:
			// pending==pending, oldest first
			if i.Updated.Before(j.Updated) {
				return 1
			}
			return -1
		case Applying, Planning, Planned, PlanQueued, ApplyQueued, Scheduled:
			// pending is after more active runs
			return 1
		default:
			// pending is before finished
			return -1
		}
	case Scheduled:
		switch j.Status {
		case Scheduled:
			// scheduled==scheduled, oldest first
			if i.Updated.Before(j.Updated) {
				return 1
			}
			return -1
		case Applying, Planning, Planned, PlanQueued, ApplyQueued:
			// scheduled is after more active runs
			return 1
		default:
			// scheduled is before pending, finished
			return -1
		}
	case PlanQueued, ApplyQueued:
		switch j.Status {
		case PlanQueued, ApplyQueued:
			// queued=queued, oldest first
			if i.Updated.Before(j.Updated) {
				return 1
			}
			return -1
		case Applying, Planning, Planned:
			// queued is after active
			return 1
		default:
			// queued is before finished
			return -1
		}
	case Planning, Applying:
		switch j.Status {
		case Planning, Applying:
			// running=running, most recent first
			if i.Updated.Before(j.Updated) {
				return -1
			}
			return 1
		default:
			// running is before everything else
			return -1
		}
	case Planned:
		switch j.Status {
		case Planned:
			// planned=planned, most recent first
			if i.Updated.Before(j.Updated) {
				return -1
			}
			return 1
		case Planning, Applying:
			// planned is after planning,applying
			return 1
		default:
			// planned is before everything else
			return -1
		}
	default: // (i: finished)
		if j.IsFinished() {
			// finished=finished (ordered by last updated desc)
			if i.Updated.Before(j.Updated) {
				return 1
			}
			return -1
		}
		// finished is after non-finished runs
		return 1
	}
}
