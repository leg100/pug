package state

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/leg100/pug/internal/resource"
)

type State struct {
	WorkspaceID      resource.ID
	Resources        map[ResourceAddress]*Resource
	Serial           int64
	TerraformVersion string
	Lineage          string
}

func newState(ws resource.Resource, r io.Reader) (*State, error) {
	// Default to a serial of -1 to indicate that there is no state yet.
	state := &State{WorkspaceID: ws.GetID(), Serial: -1}

	var file StateFile
	if err := json.NewDecoder(r).Decode(&file); err != nil {
		if errors.Is(err, io.EOF) {
			// No state, serial defaults to -1
			return state, nil
		}
		return nil, fmt.Errorf("parsing state: %w", err)
	}

	state.Serial = file.Serial
	state.TerraformVersion = file.TerraformVersion
	state.Lineage = file.Lineage

	m := make(map[ResourceAddress]*Resource)
	for _, res := range file.Resources {
		for _, instance := range res.Instances {
			// Build resource address from type, name, and optionally an ordinal
			// number if more than one instance.
			var b strings.Builder
			if res.Module != "" {
				b.WriteString(res.Module)
				b.WriteRune('.')
			}
			b.WriteString(res.Type)
			b.WriteRune('.')
			b.WriteString(res.Name)
			if instance.IndexKey != nil {
				b.WriteRune('[')
				b.WriteString(strconv.Itoa(*instance.IndexKey))
				b.WriteRune(']')
			}
			addr := ResourceAddress(b.String())
			m[addr] = newResource(ws, addr)
			if instance.Status == StateFileResourceInstanceTainted {
				m[addr].Tainted = true
			}
		}
	}
	state.Resources = m

	return state, nil
}
