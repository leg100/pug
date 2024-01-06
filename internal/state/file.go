package state

import (
	"encoding/json"
)

type (
	// StateFile is the terraform state file contents
	StateFile struct {
		FormatVersion    string `json:"format_version"`
		TerraformVersion string `json:"terraform_version"`
		Values           StateFileValues
	}

	StateFileValues struct {
		Outputs    map[string]StateFileOutput
		RootModule StateFileModule `json:"root_module"`
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
		Address ResourceAddress
		Tainted bool
	}
)

func getResourcesFromFile(f StateFile) map[ResourceAddress]*Resource {
	m := make(map[ResourceAddress]*Resource)
	return getResourcesFromStateFileModule(f.Values.RootModule, m)
}

func getResourcesFromStateFileModule(mod StateFileModule, m map[ResourceAddress]*Resource) map[ResourceAddress]*Resource {
	for _, res := range mod.Resources {
		m[res.Address] = &Resource{
			Address: res.Address,
		}
		if res.Tainted {
			m[res.Address].Status = Tainted
		}
	}
	for _, child := range mod.ChildModules {
		m = getResourcesFromStateFileModule(child, m)
	}
	return m
}
