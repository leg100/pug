package run

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
)

var (
	planChangesRegex       = regexp.MustCompile(`Plan: (\d+) to add, (\d+) to change, (\d+) to destroy.`)
	planOutputChangesRegex = regexp.MustCompile(`Changes to Outputs:`)
	planNoChangesRegex     = regexp.MustCompile(`No changes. Your infrastructure matches the configuration.`)
	applyChangesRegex      = regexp.MustCompile(`(?m)^Apply complete! Resources: (\d+) added, (\d+) changed, (\d+) destroyed.`)
)

// parsePlanReport reads the logs from `terraform plan` and detects whether
// there were any changes and produces a report of the number of resource
// changes.
func parsePlanReport(logs string) (bool, report, error) {
	raw := stripAnsi(logs)

	// No changes
	if planNoChangesRegex.MatchString(raw) {
		return false, report{}, nil
	}

	// Resource changes
	matches := planChangesRegex.FindStringSubmatch(raw)
	if matches != nil {
		adds, err := strconv.ParseInt(matches[1], 10, 0)
		if err != nil {
			return false, report{}, err
		}
		changes, err := strconv.ParseInt(matches[2], 10, 0)
		if err != nil {
			return false, report{}, err
		}
		deletions, err := strconv.ParseInt(matches[3], 10, 0)
		if err != nil {
			return false, report{}, err
		}
		return true, report{
			Additions:    int(adds),
			Changes:      int(changes),
			Destructions: int(deletions),
		}, nil
	}

	// Output changes
	if planOutputChangesRegex.MatchString(raw) {
		return true, report{}, nil
	}

	// Something went wrong
	return false, report{}, errors.New("unexpected plan output: failed to detect changes")
}

// parseApplyReport reads the logs from `terraform apply` and produces a report
// of the changes made.
func parseApplyReport(logs string) (report, error) {
	matches := applyChangesRegex.FindStringSubmatch(logs)
	if matches == nil {
		return report{}, fmt.Errorf("regexes unexpectedly did not match apply output")
	}

	adds, err := strconv.ParseInt(matches[1], 10, 0)
	if err != nil {
		return report{}, err
	}
	changes, err := strconv.ParseInt(matches[2], 10, 0)
	if err != nil {
		return report{}, err
	}
	deletions, err := strconv.ParseInt(matches[3], 10, 0)
	if err != nil {
		return report{}, err
	}

	return report{
		Additions:    int(adds),
		Changes:      int(changes),
		Destructions: int(deletions),
	}, nil
}
