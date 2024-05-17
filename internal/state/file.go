package state

import (
	"encoding/json"
)

const (
	StateFileResourceInstanceTainted StateFileResourceInstanceStatus = "tainted"
)

type (
	// StateFile is the terraform state file contents
	StateFile struct {
		Version          int
		TerraformVersion string `json:"terraform_version"`
		Serial           int64
		Lineage          string
		Outputs          map[string]StateFileOutput
		Resources        []StateFileResource
	}

	// StateFileOutput is an output in the terraform state file
	StateFileOutput struct {
		Value     json.RawMessage
		Sensitive bool
	}

	StateFileModule struct {
		Resources    []StateFileResource
		ChildModules []StateFileModule `json:"child_modules"`
	}

	StateFileResource struct {
		Name        string
		ProviderURI string `json:"provider"`
		Type        string
		Module      string
		Instances   []StateFileResourceInstance
	}

	StateFileResourceInstance struct {
		IndexKey   *int `json:"index_key"`
		Status     StateFileResourceInstanceStatus
		Attributes json.RawMessage
	}

	StateFileResourceInstanceStatus string
)
