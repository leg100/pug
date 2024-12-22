package app

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/leg100/pug/internal/app"
	"github.com/stretchr/testify/require"
)

func TestCost(t *testing.T) {
	t.Parallel()
	skipIfInfracostNotFound(t)

	tm := setupInfracostWorkspaces(t)

	// Navigate explorer cursor to first workspace
	tm.Send(tea.KeyMsg{Type: tea.KeyDown})

	// Calculate cost for all four workspaces
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlA})
	tm.Type("$")

	// Wait for infracost task to produce overall total. Each workspace in the
	// explorer should also have a cost alongside it.
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "cost") &&
			strings.Contains(s, "exited $2621.90") &&
			matchPattern(t, `OVERALL TOTAL.*\$2\,621\.90`, s) &&
			strings.Contains(s, `default 0 $87.12`) &&
			strings.Contains(s, `dev 0 $2360.55`) &&
			strings.Contains(s, `default 0 $87.12`) &&
			strings.Contains(s, `default 0 $87.12`)
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

	// Expect three modules in tree
	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "├ 󰠱 a") &&
			strings.Contains(s, "├ 󰠱 b") &&
			strings.Contains(s, "└ 󰠱 c")
	})

	// Select all modules and init
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlA})
	tm.Type("i")

	waitFor(t, tm, func(s string) bool {
		return strings.Contains(s, "init 3/3") &&
			matchPattern(t, `modules/a.*init.*exited`, s) &&
			matchPattern(t, `modules/b.*init.*exited`, s) &&
			matchPattern(t, `modules/c.*init.*exited`, s) &&
			strings.Contains(s, "󰠱 3  4")
	})

	// Go back to explorer
	tm.Type("0")

	// Clear selection
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlBackslash})

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
