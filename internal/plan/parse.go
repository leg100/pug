package plan

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"

	"github.com/leg100/pug/internal"
)

var (
	planChangesRegex       = regexp.MustCompile(`Plan: (\d+) to add, (\d+) to change, (\d+) to destroy.`)
	planOutputChangesRegex = regexp.MustCompile(`Changes to Outputs:`)
	noChangesRegex         = regexp.MustCompile(`No changes. Your infrastructure matches the configuration.`)
	applyChangesRegex      = regexp.MustCompile(`(?m)^Apply complete! Resources: (\d+) added, (\d+) changed, (\d+) destroyed.`)
	destroyChangesRegex    = regexp.MustCompile(`Destroy complete! Resources: (\d+) destroyed.`)
)

// parsePlanReport reads the logs from `terraform plan` and detects whether
// there were any changes and produces a report of the number of resource
// changes.
//
// TODO: parse bytes intead, to skip re-allocation
func parsePlanReport(logs string) (bool, Report, error) {
	raw := internal.StripAnsi(logs)

	// No changes
	if noChangesRegex.MatchString(raw) {
		return false, Report{}, nil
	}

	// Resource changes
	matches := planChangesRegex.FindStringSubmatch(raw)
	if matches != nil {
		adds, err := strconv.ParseInt(matches[1], 10, 0)
		if err != nil {
			return false, Report{}, err
		}
		changes, err := strconv.ParseInt(matches[2], 10, 0)
		if err != nil {
			return false, Report{}, err
		}
		deletions, err := strconv.ParseInt(matches[3], 10, 0)
		if err != nil {
			return false, Report{}, err
		}
		return true, Report{
			Additions:    int(adds),
			Changes:      int(changes),
			Destructions: int(deletions),
		}, nil
	}

	// Output changes
	if planOutputChangesRegex.MatchString(raw) {
		return true, Report{}, nil
	}

	// Something went wrong
	return false, Report{}, errors.New("unexpected plan output: failed to detect changes")
}

// parseApplyReport reads the logs from `terraform apply` and produces a report
// of the changes made.
//
// TODO: parse bytes intead, to skip re-allocation
func parseApplyReport(logs string) (Report, error) {
	raw := internal.StripAnsi(logs)

	if noChangesRegex.MatchString(raw) {
		// No changes
		return Report{}, nil
	}
	if matches := destroyChangesRegex.FindStringSubmatch(raw); len(matches) > 1 {
		// Destroy op
		deletions, err := strconv.Atoi(matches[1])
		if err != nil {
			return Report{}, err
		}
		return Report{Destructions: deletions}, nil
	}
	if matches := applyChangesRegex.FindStringSubmatch(raw); len(matches) > 3 {
		// Apply op
		adds, err := strconv.Atoi(matches[1])
		if err != nil {
			return Report{}, err
		}
		changes, err := strconv.Atoi(matches[2])
		if err != nil {
			return Report{}, err
		}
		deletions, err := strconv.Atoi(matches[3])
		if err != nil {
			return Report{}, err
		}
		return Report{
			Additions:    adds,
			Changes:      changes,
			Destructions: deletions,
		}, nil
	}
	return Report{}, fmt.Errorf("regexes unexpectedly did not match apply output")
}
