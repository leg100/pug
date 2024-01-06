package run

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
	planNoChangesRegex     = regexp.MustCompile(`No changes. Your infrastructure matches the configuration.`)
	applyChangesRegex      = regexp.MustCompile(`(?m)^Apply complete! Resources: (\d+) added, (\d+) changed, (\d+) destroyed.`)
)

// parsePlanReport reads the logs from `terraform plan` and detects whether
// there were any changes and produces a report of the number of resource
// changes.
func parsePlanReport(logs string) (bool, Report, error) {
	raw := internal.StripAnsi(logs)

	// No changes
	if planNoChangesRegex.MatchString(raw) {
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
func parseApplyReport(logs string) (Report, error) {
	matches := applyChangesRegex.FindStringSubmatch(logs)
	if matches == nil {
		return Report{}, fmt.Errorf("regexes unexpectedly did not match apply output")
	}

	adds, err := strconv.ParseInt(matches[1], 10, 0)
	if err != nil {
		return Report{}, err
	}
	changes, err := strconv.ParseInt(matches[2], 10, 0)
	if err != nil {
		return Report{}, err
	}
	deletions, err := strconv.ParseInt(matches[3], 10, 0)
	if err != nil {
		return Report{}, err
	}

	return Report{
		Additions:    int(adds),
		Changes:      int(changes),
		Destructions: int(deletions),
	}, nil
}
