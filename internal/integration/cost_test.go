package app

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/leg100/pug/internal/app"
	"github.com/stretchr/testify/require"
)

func TestCost(t *testing.T) {
	tm := setupInfracostWorkspaces(t)

	// Calculate cost for all four workspaces
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlA})
	tm.Type("$")

	// Wait for infracost task to produce overall total
	waitFor(t, tm, func(s string) bool {
		return matchPattern(t, `Task.*cost.*exited`, s) &&
			matchPattern(t, `OVERALL TOTAL.*\$264.25`, s)
	})

	// Go back to workspace listing
	tm.Type("w")

	// Each workspace should now have a cost.
	waitFor(t, tm, func(s string) bool {
		return matchPattern(t, `modules/a.*default.*\$8.392`, s) &&
			matchPattern(t, `modules/a.*dev.*\$239.072`, s) &&
			matchPattern(t, `modules/b.*default.*\$8.392`, s) &&
			matchPattern(t, `modules/c.*default.*\$8.392`, s)
	})
}

func setupInfracostWorkspaces(t *testing.T) *testModel {
	responses := map[string][]byte{
		"t3.micro":    []byte(`[{"data":{"products":[{"prices":[{"priceHash":"2f1bc092c9e34dc084a4d96d19ef47ca-d2c98780d7b6e36641b521f1f8145c6f","USD":"0.0104"}]}]}},{"data":{"products":[{"prices":[{"priceHash":"ccdf11d8e4c0267d78a19b6663a566c1-e8e892be2fbd1c8f42fd6761ad8977d8","USD":"0.05"}]}]}},{"data":{"products":[{"prices":[{"priceHash":"efa8e70ebe004d2e9527fd30d50d09b2-ee3dd7e4624338037ca6fea0933a662f","USD":"0.1"}]}]}}]}]}}]`),
		"m7g.2xlarge": []byte(`[{"data":{"products":[{"prices":[{"priceHash":"9e451d8e8608d394a391b3709c9c8099-d2c98780d7b6e36641b521f1f8145c6f","USD":"0.3264"}]}]}},{"data":{"products":[{"prices":[{"priceHash":"efa8e70ebe004d2e9527fd30d50d09b2-ee3dd7e4624338037ca6fea0933a662f","USD":"0.1"}]}]}}]`),
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
	})

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
