package workspace

import (
	"errors"
	"io"
	"regexp"

	"github.com/leg100/pug/internal"
)

var (
	overallCostRegex = regexp.MustCompile(`OVERALL TOTAL.*(\$\d+.\d+)`)
)

func parseInfracostOutput(r io.Reader) (string, error) {
	out, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}
	raw := internal.StripAnsi(string(out))

	if matches := overallCostRegex.FindStringSubmatch(raw); len(matches) > 1 {
		return matches[1], nil
	}
	return "", errors.New("failed to parsed overall cost from infracost output")
}
