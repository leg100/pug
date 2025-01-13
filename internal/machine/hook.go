// Copyright (c) The OpenTofu Authors
// SPDX-License-Identifier: MPL-2.0
// Copyright (c) 2023 HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package machine

import (
	"fmt"
	"time"
)

type Hook interface {
	HookType() MessageType
	String() string
}

// ApplyStart: triggered by PreApply hook
type applyStart struct {
	Resource   ResourceAddr `json:"resource"`
	Action     ChangeAction `json:"action"`
	IDKey      string       `json:"id_key,omitempty"`
	IDValue    string       `json:"id_value,omitempty"`
	actionVerb string
}

var _ Hook = (*applyStart)(nil)

func (h *applyStart) HookType() MessageType {
	return MessageApplyStart
}

func (h *applyStart) String() string {
	var id string
	if h.IDKey != "" && h.IDValue != "" {
		id = fmt.Sprintf(" [%s=%s]", h.IDKey, h.IDValue)
	}
	return fmt.Sprintf("%s: %s...%s", h.Resource.Addr, h.actionVerb, id)
}

// ApplyProgress: currently triggered by a timer started on PreApply. In
// future, this might also be triggered by provider progress reporting.
type applyProgress struct {
	Resource   ResourceAddr `json:"resource"`
	Action     ChangeAction `json:"action"`
	Elapsed    float64      `json:"elapsed_seconds"`
	actionVerb string
	elapsed    time.Duration
}

var _ Hook = (*applyProgress)(nil)

func (h *applyProgress) HookType() MessageType {
	return MessageApplyProgress
}

func (h *applyProgress) String() string {
	return fmt.Sprintf("%s: Still %s... [%s elapsed]", h.Resource.Addr, h.actionVerb, h.elapsed)
}

// ApplyComplete: triggered by PostApply hook
type applyComplete struct {
	Resource   ResourceAddr `json:"resource"`
	Action     ChangeAction `json:"action"`
	IDKey      string       `json:"id_key,omitempty"`
	IDValue    string       `json:"id_value,omitempty"`
	Elapsed    float64      `json:"elapsed_seconds"`
	actionNoun string
	elapsed    time.Duration
}

var _ Hook = (*applyComplete)(nil)

func (h *applyComplete) HookType() MessageType {
	return MessageApplyComplete
}

func (h *applyComplete) String() string {
	var id string
	if h.IDKey != "" && h.IDValue != "" {
		id = fmt.Sprintf(" [%s=%s]", h.IDKey, h.IDValue)
	}
	return fmt.Sprintf("%s: %s complete after %s%s", h.Resource.Addr, h.actionNoun, h.elapsed, id)
}

// ApplyErrored: triggered by PostApply hook on failure. This will be followed
// by diagnostics when the apply finishes.
type applyErrored struct {
	Resource   ResourceAddr `json:"resource"`
	Action     ChangeAction `json:"action"`
	Elapsed    float64      `json:"elapsed_seconds"`
	actionNoun string
	elapsed    time.Duration
}

var _ Hook = (*applyErrored)(nil)

func (h *applyErrored) HookType() MessageType {
	return MessageApplyErrored
}

func (h *applyErrored) String() string {
	return fmt.Sprintf("%s: %s errored after %s", h.Resource.Addr, h.actionNoun, h.elapsed)
}

// ProvisionStart: triggered by PreProvisionInstanceStep hook
type provisionStart struct {
	Resource    ResourceAddr `json:"resource"`
	Provisioner string       `json:"provisioner"`
}

var _ Hook = (*provisionStart)(nil)

func (h *provisionStart) HookType() MessageType {
	return MessageProvisionStart
}

func (h *provisionStart) String() string {
	return fmt.Sprintf("%s: Provisioning with '%s'...", h.Resource.Addr, h.Provisioner)
}

// ProvisionProgress: triggered by ProvisionOutput hook
type provisionProgress struct {
	Resource    ResourceAddr `json:"resource"`
	Provisioner string       `json:"provisioner"`
	Output      string       `json:"output"`
}

var _ Hook = (*provisionProgress)(nil)

func (h *provisionProgress) HookType() MessageType {
	return MessageProvisionProgress
}

func (h *provisionProgress) String() string {
	return fmt.Sprintf("%s: (%s): %s", h.Resource.Addr, h.Provisioner, h.Output)
}

// ProvisionComplete: triggered by PostProvisionInstanceStep hook
type provisionComplete struct {
	Resource    ResourceAddr `json:"resource"`
	Provisioner string       `json:"provisioner"`
}

var _ Hook = (*provisionComplete)(nil)

func (h *provisionComplete) HookType() MessageType {
	return MessageProvisionComplete
}

func (h *provisionComplete) String() string {
	return fmt.Sprintf("%s: (%s) Provisioning complete", h.Resource.Addr, h.Provisioner)
}

// ProvisionErrored: triggered by PostProvisionInstanceStep hook on failure.
// This will be followed by diagnostics when the apply finishes.
type provisionErrored struct {
	Resource    ResourceAddr `json:"resource"`
	Provisioner string       `json:"provisioner"`
}

var _ Hook = (*provisionErrored)(nil)

func (h *provisionErrored) HookType() MessageType {
	return MessageProvisionErrored
}

func (h *provisionErrored) String() string {
	return fmt.Sprintf("%s: (%s) Provisioning errored", h.Resource.Addr, h.Provisioner)
}

// RefreshStart: triggered by PreRefresh hook
type refreshStart struct {
	Resource ResourceAddr `json:"resource"`
	IDKey    string       `json:"id_key,omitempty"`
	IDValue  string       `json:"id_value,omitempty"`
}

var _ Hook = (*refreshStart)(nil)

func (h *refreshStart) HookType() MessageType {
	return MessageRefreshStart
}

func (h *refreshStart) String() string {
	var id string
	if h.IDKey != "" && h.IDValue != "" {
		id = fmt.Sprintf(" [%s=%s]", h.IDKey, h.IDValue)
	}
	return fmt.Sprintf("%s: Refreshing state...%s", h.Resource.Addr, id)
}

// RefreshComplete: triggered by PostRefresh hook
type refreshComplete struct {
	Resource ResourceAddr `json:"resource"`
	IDKey    string       `json:"id_key,omitempty"`
	IDValue  string       `json:"id_value,omitempty"`
}

var _ Hook = (*refreshComplete)(nil)

func (h *refreshComplete) HookType() MessageType {
	return MessageRefreshComplete
}

func (h *refreshComplete) String() string {
	var id string
	if h.IDKey != "" && h.IDValue != "" {
		id = fmt.Sprintf(" [%s=%s]", h.IDKey, h.IDValue)
	}
	return fmt.Sprintf("%s: Refresh complete%s", h.Resource.Addr, id)
}
