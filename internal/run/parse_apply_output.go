package run

import (
	"fmt"
	"regexp"
	"strconv"
)

var applyChangesRegex = regexp.MustCompile(`(?m)^Apply complete! Resources: (\d+) added, (\d+) changed, (\d+) destroyed.`)

func parseApplyOutput(output string) (report, error) {
	matches := applyChangesRegex.FindStringSubmatch(output)
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
