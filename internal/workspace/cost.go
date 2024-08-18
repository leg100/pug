package workspace

import (
	"errors"
	"io"
	"regexp"

	"github.com/leg100/pug/internal"
	"github.com/leg100/pug/internal/module"
	"github.com/leg100/pug/internal/resource"
	"gopkg.in/yaml.v3"
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

func generateInfracostConfig(modulesAndWorkspaces ...resource.Resource) ([]byte, error) {
	cfg := infracostConfig{Version: "0.1"}
	cfg.Projects = make([]infracostProjectConfig, len(modulesAndWorkspaces))

	for i, res := range modulesAndWorkspaces {
		switch res := res.(type) {
		case *module.Module:
			cfg.Projects[i] = infracostProjectConfig{
				Path: res.Path,
			}
		case *Workspace:
			cfg.Projects[i] = infracostProjectConfig{
				Path:               res.ModulePath(),
				TerraformWorkspace: res.Name,
			}
			if fname, ok := res.VarsFile(); ok {
				cfg.Projects[i].TerraformVarFiles = []string{fname}
			}
		}
	}

	return yaml.Marshal(cfg)
}
