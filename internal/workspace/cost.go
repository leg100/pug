package workspace

import (
	"encoding/json"

	"github.com/leg100/pug/internal"
	"gopkg.in/yaml.v3"
)

type infracostConfig struct {
	Version  string
	Projects []infracostProjectConfig
}

type infracostProjectConfig struct {
	Path               string
	Name               string   `yaml:",omitempty"`
	TerraformWorkspace string   `yaml:"terraform_workspace,omitempty"`
	TerraformVarFiles  []string `yaml:"terraform_var_files,omitempty"`
}

func generateCostConfig(workdir internal.Workdir, workspaces ...*Workspace) ([]byte, error) {
	cfg := infracostConfig{Version: "0.1"}
	cfg.Projects = make([]infracostProjectConfig, len(workspaces))

	for i, ws := range workspaces {
		cfg.Projects[i] = infracostProjectConfig{
			Path:               ws.ModulePath(),
			TerraformWorkspace: ws.Name,
		}
		if fname, ok := ws.VarsFile(workdir); ok {
			cfg.Projects[i].TerraformVarFiles = []string{fname}
		}
	}

	return yaml.Marshal(cfg)
}

type infracostBreakdown struct {
	Version  string
	Projects []infracostBreakdownProject
}

type infracostBreakdownProject struct {
	Metadata  infracostBreakdownProjectMetadata
	Breakdown infracostBreakdownBreakdown
}

type infracostBreakdownProjectMetadata struct {
	TerraformModulePath string `json:"terraformModulePath"`
	TerraformWorkspace  string `json:"terraformWorkspace"`
}

type infracostBreakdownBreakdown struct {
	TotalMonthlyCost string `json:"totalMonthlyCost"`
}

type breakdownResult struct {
	path      string
	workspace string
	cost      string
}

func parseBreakdown(jsonPayload []byte) ([]breakdownResult, error) {
	var breakdown infracostBreakdown
	if err := json.Unmarshal(jsonPayload, &breakdown); err != nil {
		return nil, err
	}
	results := make([]breakdownResult, len(breakdown.Projects))
	for i, proj := range breakdown.Projects {
		results[i] = breakdownResult{
			path:      proj.Metadata.TerraformModulePath,
			workspace: proj.Metadata.TerraformWorkspace,
			cost:      proj.Breakdown.TotalMonthlyCost,
		}
	}
	return results, nil
}
