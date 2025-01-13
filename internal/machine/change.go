// Copyright (c) The OpenTofu Authors
// SPDX-License-Identifier: MPL-2.0
// Copyright (c) 2023 HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package machine

import (
	"fmt"
)

type ResourceInstanceChange struct {
	Resource         ResourceAddr  `json:"resource"`
	PreviousResource *ResourceAddr `json:"previous_resource,omitempty"`
	Action           ChangeAction  `json:"action"`
	Reason           ChangeReason  `json:"reason,omitempty"`
	Importing        *Importing    `json:"importing,omitempty"`
	GeneratedConfig  string        `json:"generated_config,omitempty"`
}

func (c *ResourceInstanceChange) String() string {
	return fmt.Sprintf("%s: Plan to %s", c.Resource.Addr, c.Action)
}

type ChangeAction string

const (
	ActionNoOp    ChangeAction = "noop"
	ActionMove    ChangeAction = "move"
	ActionCreate  ChangeAction = "create"
	ActionRead    ChangeAction = "read"
	ActionUpdate  ChangeAction = "update"
	ActionReplace ChangeAction = "replace"
	ActionDelete  ChangeAction = "delete"
	ActionImport  ChangeAction = "import"
	ActionForget  ChangeAction = "remove"
)

type ChangeReason string

const (
	ReasonNone               ChangeReason = ""
	ReasonTainted            ChangeReason = "tainted"
	ReasonRequested          ChangeReason = "requested"
	ReasonReplaceTriggeredBy ChangeReason = "replace_triggered_by"
	ReasonCannotUpdate       ChangeReason = "cannot_update"
	ReasonUnknown            ChangeReason = "unknown"

	ReasonDeleteBecauseNoResourceConfig ChangeReason = "delete_because_no_resource_config"
	ReasonDeleteBecauseWrongRepetition  ChangeReason = "delete_because_wrong_repetition"
	ReasonDeleteBecauseCountIndex       ChangeReason = "delete_because_count_index"
	ReasonDeleteBecauseEachKey          ChangeReason = "delete_because_each_key"
	ReasonDeleteBecauseNoModule         ChangeReason = "delete_because_no_module"
	ReasonDeleteBecauseNoMoveTarget     ChangeReason = "delete_because_no_move_target"
	ReasonReadBecauseConfigUnknown      ChangeReason = "read_because_config_unknown"
	ReasonReadBecauseDependencyPending  ChangeReason = "read_because_dependency_pending"
	ReasonReadBecauseCheckNested        ChangeReason = "read_because_check_nested"
)
