// Copyright (c) The OpenTofu Authors
// SPDX-License-Identifier: MPL-2.0
// Copyright (c) 2023 HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package machine

import (
	"encoding/json"
)

// Function is a description of the JSON representation of the signature of
// a function callable from the OpenTofu language.
type Function struct {
	// Name is the leaf name of the function, without any namespace prefix.
	Name string `json:"name"`

	Params        []FunctionParam `json:"params"`
	VariadicParam *FunctionParam  `json:"variadic_param,omitempty"`

	// ReturnType is type constraint which is a static approximation of the
	// possibly-dynamic return type of the function.
	ReturnType json.RawMessage `json:"return_type"`

	Description     string `json:"description,omitempty"`
	DescriptionKind string `json:"description_kind,omitempty"`
}

// FunctionParam represents a single parameter to a function, as represented
// by type Function.
type FunctionParam struct {
	// Name is a name for the function which is used primarily for
	// documentation purposes, because function arguments are positional
	// and therefore don't appear directly in configuration source code.
	Name string `json:"name"`

	// Type is a type constraint which is a static approximation of the
	// possibly-dynamic type of the parameter. Particular functions may
	// have additional requirements that a type constraint alone cannot
	// represent.
	Type json.RawMessage `json:"type"`

	// Maybe some of the other fields in function.Parameter would be
	// interesting to describe here too, but we'll wait to see if there
	// is a use-case first.

	Description     string `json:"description,omitempty"`
	DescriptionKind string `json:"description_kind,omitempty"`
}
