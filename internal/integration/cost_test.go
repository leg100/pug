package app

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"
	"github.com/leg100/pug/internal/app"
	"github.com/stretchr/testify/require"
)

func TestCost(t *testing.T) {
	t.Parallel()
	skipIfInfracostNotFound(t)

	tm := setupInfracostWorkspaces(t)

	// Calculate cost for all four workspaces
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlA})
	tm.Type("$")

	// Wait for infracost task to produce overall total
	waitFor(t, tm, func(s string) bool {
		return matchPattern(t, `Task.*cost.*exited`, s) &&
			matchPattern(t, `OVERALL TOTAL.*\$2\,621\.90`, s)
	})

	// Go back to workspace listing
	tm.Type("w")

	// Each workspace should now have a cost.
	waitFor(t, tm, func(s string) bool {
		return matchPattern(t, `modules/a.*default.*\$87.116912`, s) &&
			matchPattern(t, `modules/a.*dev.*\$2360.547736`, s) &&
			matchPattern(t, `modules/b.*default.*\$87.116912`, s) &&
			matchPattern(t, `modules/c.*default.*\$87.116912`, s)
	})
}

func setupInfracostWorkspaces(t *testing.T) *testModel {
	responses := map[string][]byte{
		"n2-standard-2":  []byte(`[{"data":{"products":[{"prices":[{"priceHash":"3460e5656b29ac302574c1c49d98a379-66d0d770bee368b4f2a8f2f597eeb417","USD":"0.097118"}]}]}},{"data":{"products":[{"prices":[{"priceHash":"4e58b7b536714dfce35b3050caa6034b-af6a951f170fc579633ad2c8f86a9dca","USD":"0.04"}]}]}},{"data":{"products":[{"prices":[{"priceHash":"dae3672d3f7605d4e5c6d48aa342d66c-57bc5d148491a8381abaccb21ca6b4e9","USD":"0.08"}]}]}}]`),
		"n1-standard-96": []byte(`[{"data":{"products":[{"prices":[{"priceHash":"84f8a08589f2331eac14c963f98e7f73-66d0d770bee368b4f2a8f2f597eeb417","USD":"4.559976"}]}]}},{"data":{"products":[{"prices":[{"priceHash":"4e58b7b536714dfce35b3050caa6034b-af6a951f170fc579633ad2c8f86a9dca","USD":"0.04"}]}]}},{"data":{"products":[{"prices":[{"priceHash":"dae3672d3f7605d4e5c6d48aa342d66c-57bc5d148491a8381abaccb21ca6b4e9","USD":"0.08"}]}]}}]`),
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		for instanceType, res := range responses {
			if strings.Contains(string(body), instanceType) {
				w.Write(res)
				return
			}
		}
	}))
	t.Cleanup(srv.Close)

	tm := setup(t, "./testdata/workspaces_that_cost_money", withInfracostEnvs(srv.URL))

	// Expect three modules to be listed
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "modules/a") &&
			strings.Contains(s, "modules/b") &&
			strings.Contains(s, "modules/c")
	})

	// Select all modules and init
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlA})
	tm.Type("i")
	waitFor(t, tm, func(s string) bool {
		return matchPattern(t, "TaskGroup.*init", s) &&
			matchPattern(t, `modules/a.*exited`, s) &&
			matchPattern(t, `modules/b.*exited`, s) &&
			matchPattern(t, `modules/c.*exited`, s)
	}, teatest.WithDuration(time.Second*30))

	// Go to workspace listing
	tm.Type("w")

	// Wait for all four workspaces to be listed.
	waitFor(t, tm, func(s string) bool {
		return matchPattern(t, `modules/a.*default`, s) &&
			matchPattern(t, `modules/a.*dev`, s) &&
			matchPattern(t, `modules/b.*default`, s) &&
			matchPattern(t, `modules/c.*default`, s)
	})

	return tm
}

func withInfracostEnvs(pricingEndpoint string) configOption {
	return func(cfg *app.Config) {
		cfg.Envs = []string{
			fmt.Sprintf("PRICING_API_ENDPOINT=%s", pricingEndpoint),
			"INFRACOST_API_KEY=ico-abc",
		}
	}
}

func skipIfInfracostNotFound(t *testing.T) {
	if _, err := exec.LookPath("infracost"); err != nil {
		t.Skip("skipping test: infracost not found")
	}
}
