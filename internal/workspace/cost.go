package workspace

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/google/uuid"
	"github.com/leg100/pug/internal"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/task"
	"gopkg.in/yaml.v3"
)

type costTaskSpecCreator struct {
	*Service
}

// Cost creates a task that retrieves a breakdown of the costs of the
// infrastructure deployed by the workspace.
func (s *costTaskSpecCreator) Cost(workspaceIDs ...resource.ID) (task.Spec, error) {
	if len(workspaceIDs) == 0 {
		return task.Spec{}, errors.New("no workspaces specified")
	}
	workspaces := make([]*Workspace, len(workspaceIDs))
	for i, id := range workspaceIDs {
		ws, err := s.Get(id)
		if err != nil {
			return task.Spec{}, err
		}
		workspaces[i] = ws
	}
	var (
		configPath    string
		breakdownPath string
	)
	{
		// generate unique names for temporary files
		id := uuid.New()
		configPath = filepath.Join(s.datadir, fmt.Sprintf("cost-%s.yaml", id.String()))
		breakdownPath = filepath.Join(s.datadir, fmt.Sprintf("breakdown-%s.json", id.String()))
	}
	{
		// generate config for infracost
		configBody, err := generateCostConfig(s.workdir, workspaces...)
		if err != nil {
			return task.Spec{}, err
		}
		if err := os.WriteFile(configPath, configBody, 0o644); err != nil {
			return task.Spec{}, err
		}
	}
	return task.Spec{
		Execution: task.Execution{
			Program: "infracost",
			Args: []string{
				"breakdown",
				"--config-file", configPath,
				"--format", "json",
				"--out-file", breakdownPath,
			},
		},
		AdditionalExecution: &task.Execution{
			Program: "infracost",
			Args: []string{
				"output",
				"--format", "table",
				"--path", breakdownPath,
			},
		},
		Blocking:    true,
		Description: "cost",
		BeforeExited: func(*task.Task) (task.Summary, error) {
			// Parse JSON output and update workspaces.
			breakdown, err := os.ReadFile(breakdownPath)
			if err != nil {
				return nil, err
			}
			result, err := parseBreakdown(breakdown)
			if err != nil {
				return nil, err
			}
			for _, result := range result.projects {
				ws, err := s.GetByName(result.path, result.workspace)
				if err != nil {
					return nil, err
				}
				_, err = s.table.Update(ws.ID, func(existing *Workspace) error {
					existing.Cost = result.cost
					return nil
				})
				if err != nil {
					return nil, err
				}
			}
			return CostSummary(result.total), nil
		},
		AfterFinish: func(*task.Task) {
			os.Remove(configPath)
			os.Remove(breakdownPath)
		},
	}, nil
}

type CostSummary float64

func (c CostSummary) String() string {
	return fmt.Sprintf("$%.2f", c)
}

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
			Path:               ws.ModulePath,
			TerraformWorkspace: ws.Name,
		}
		if fname, ok := ws.VarsFile(workdir); ok {
			cfg.Projects[i].TerraformVarFiles = []string{fname}
		}
	}

	return yaml.Marshal(cfg)
}

type infracostBreakdown struct {
	Version          string
	Projects         []infracostBreakdownProject
	TotalMonthlyCost string `json:"totalMonthlyCost"`
}

type infracostBreakdownProject struct {
	Metadata  infracostBreakdownProjectMetadata
	Breakdown infracostBreakdownProjectBreakdown
}

type infracostBreakdownProjectMetadata struct {
	TerraformModulePath string `json:"terraformModulePath"`
	TerraformWorkspace  string `json:"terraformWorkspace"`
}

type infracostBreakdownProjectBreakdown struct {
	TotalMonthlyCost string `json:"totalMonthlyCost"`
}

type breakdownResult struct {
	projects []breakdownResultProject
	total    float64
}

type breakdownResultProject struct {
	path      string
	workspace string
	cost      float64
}

func parseBreakdown(jsonPayload []byte) (breakdownResult, error) {
	var breakdown infracostBreakdown
	if err := json.Unmarshal(jsonPayload, &breakdown); err != nil {
		return breakdownResult{}, err
	}
	// Parse overall total cost
	total, err := strconv.ParseFloat(breakdown.TotalMonthlyCost, 64)
	if err != nil {
		return breakdownResult{}, err
	}
	// Parse per-project costs
	result := breakdownResult{total: total}
	result.projects = make([]breakdownResultProject, len(breakdown.Projects))
	for i, proj := range breakdown.Projects {
		result.projects[i] = breakdownResultProject{
			path:      proj.Metadata.TerraformModulePath,
			workspace: proj.Metadata.TerraformWorkspace,
		}
		cost, err := strconv.ParseFloat(proj.Breakdown.TotalMonthlyCost, 64)
		if err != nil {
			return breakdownResult{}, err
		}
		result.projects[i].cost = cost
	}
	return result, nil
}
