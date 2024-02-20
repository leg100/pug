package run

import "testing"

func TestRun(t *testing.T) {
	r := newRun()
	t.Log(r.id)
	t.Log(len(r.id))
}
