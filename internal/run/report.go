package run

import (
	"fmt"
)

// report reports a summary of additions, changes, and deletions of
// resources in a plan or an apply.
type report struct {
	Additions    int `json:"additions"`
	Changes      int `json:"changes"`
	Destructions int `json:"destructions"`
}

func (r report) HasChanges() bool {
	return r != report{}
}

func (r report) String() string {
	// \u2212 is a proper minus sign; an ascii hyphen is too narrow (in the
	// default github font at least) and looks incongruous alongside
	// the wider '+' and '~' characters.
	return fmt.Sprintf("+%d/~%d/\u2212%d", r.Additions, r.Changes, r.Destructions)
}
