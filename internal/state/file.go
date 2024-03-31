package state

import (
	"encoding/json"
	"strings"
)

type (
	// StateFile is the terraform state file contents
	StateFile struct {
		Version          int
		TerraformVersion string `json:"terraform_version"`
		Serial           int64
		Lineage          string
		Outputs          map[string]StateFileOutput
		FileResources    []StateFileResource `json:"resources"`
	}

	// StateFileOutput is an output in the terraform state file
	StateFileOutput struct {
		Value     json.RawMessage
		Sensitive bool
	}

	StateFileResource struct {
		Name        string
		ProviderURI string `json:"provider"`
		Type        string
		Module      string
	}
)

func (f StateFile) Resources() map[ResourceAddress]*Resource {
	resources := make(map[ResourceAddress]*Resource, len(f.FileResources))
	for _, fr := range f.FileResources {
		r := newResource(fr)
		resources[r.Address] = r
	}
	return resources
}

func (r StateFileResource) ModuleName() string {
	if r.Module == "" {
		return "root"
	}
	return strings.TrimPrefix(r.Module, "module.")
}

// Type determines the HCL type of the output value
func (r StateFileOutput) Type() (string, error) {
	var dst any
	if err := json.Unmarshal(r.Value, &dst); err != nil {
		return "", err
	}

	var typ string
	switch dst.(type) {
	case bool:
		typ = "bool"
	case float64:
		typ = "number"
	case string:
		typ = "string"
	case []any:
		typ = "tuple"
	case map[string]any:
		typ = "object"
	case nil:
		typ = "null"
	default:
		typ = "unknown"
	}
	return typ, nil
}

func (r StateFileOutput) StringValue() string {
	return string(r.Value)
}
