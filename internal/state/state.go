package state

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strconv"
	"strings"

	"github.com/leg100/pug/internal/resource"
)

type State struct {
	resource.Common

	WorkspaceID      resource.ID
	Resources        map[ResourceAddress]*Resource
	Serial           int64
	TerraformVersion string
	Lineage          string
}

func newState(ws resource.Resource, r io.Reader) (*State, error) {
	// Default to a serial of -1 to indicate that there is no state yet.
	state := &State{
		Common:      resource.New(resource.State, ws),
		WorkspaceID: ws.GetID(),
		Serial:      -1,
	}

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
			if res.Mode == StateFileResourceDataMode {
				b.WriteString("data.")
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
			var err error
			m[addr], err = newResource(state, addr, instance.Attributes)
			if err != nil {
				return nil, fmt.Errorf("decoding resource %s: %w", addr, err)
			}
			if instance.Status == StateFileResourceInstanceTainted {
				m[addr].Tainted = true
			}
		}
	}
	state.Resources = m

	return state, nil
}

func (s *State) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Int("resources", len(s.Resources)),
		slog.Int64("serial", s.Serial),
	)
}
