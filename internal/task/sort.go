package task

// ByState implements sort.Interface for []*Task based on state, with the
// following ordering:
//
// 1. running (ordered by last updated desc)
// 2. queued (ordered by last updated asc)
// 3. pending (ordered by last updated asc)
// 4. finished (ordered by last updated desc)
func ByState(i, j *Task) int {
	switch i.State {
	case Pending:
		switch j.State {
		case Pending:
			// pending==pending, ordered by last udpated desc
			if i.Updated.Before(j.Updated) {
				return -1
			}
			return 1
		case Queued, Running:
			// pending is after queued and running
			return 1
		default:
			// pending is before finished
			return -1
		}
	case Queued:
		switch j.State {
		case Queued:
			// queued=queued (ordered by last updated desc)
			if i.Updated.After(j.Updated) {
				return -1
			}
			return 1
		case Running:
			// queued is after running
			return 1
		default:
			// queued is before pending and finished
			return -1
		}
	case Running:
		switch j.State {
		case Running:
			// running=running (ordered by last updated asc)
			if i.Updated.Before(j.Updated) {
				return -1
			}
			return 1
		default:
			// running is before pending, queued, and finished
			return -1
		}
	default:
		switch j.State {
		case Pending, Queued, Running:
			// finished is after pending, queued, and running
			return 1
		default:
			// finished=finished (ordered by last updated desc)
			if i.Updated.Before(j.Updated) {
				return 1
			}
			return -1
		}
	}
}
