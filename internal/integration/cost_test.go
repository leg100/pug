package app

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCost(t *testing.T) {
	res := []byte(`[{"data":{"products":[{"prices":[{"priceHash":"2f1bc092c9e34dc084a4d96d19ef47ca-d2c98780d7b6e36641b521f1f8145c6f","USD":"0.0104"}]}]}},{"data":{"products":[{"prices":[{"priceHash":"ccdf11d8e4c0267d78a19b6663a566c1-e8e892be2fbd1c8f42fd6761ad8977d8","USD":"0.05"}]}]}},{"data":{"products":[{"prices":[{"priceHash":"efa8e70ebe004d2e9527fd30d50d09b2-ee3dd7e4624338037ca6fea0933a662f","USD":"0.1"}]}]}}]`)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(res)
	}))
	t.Cleanup(srv.Close)

	cmd := exec.Command("infracost", "breakdown", "--path", "./testdata/ec2_instance/")
	cmd.Env = []string{
		fmt.Sprintf("PRICING_API_ENDPOINT=%s", srv.URL),
		"INFRACOST_API_KEY=ico-abc",
	}
	cmd.Stdout = &testLogger{t}
	cmd.Stderr = &testLogger{t}
	err := cmd.Run()
	assert.NoError(t, err)
}
