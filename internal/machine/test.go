// Copyright (c) The OpenTofu Authors
// SPDX-License-Identifier: MPL-2.0
// Copyright (c) 2023 HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package machine

type TestSuiteAbstract map[string][]string

type TestStatus string

type TestFileStatus struct {
	Path   string     `json:"path"`
	Status TestStatus `json:"status"`
}

type TestRunStatus struct {
	Path   string     `json:"path"`
	Run    string     `json:"run"`
	Status TestStatus `json:"status"`
}

type TestSuiteSummary struct {
	Status  TestStatus `json:"status"`
	Passed  int        `json:"passed"`
	Failed  int        `json:"failed"`
	Errored int        `json:"errored"`
	Skipped int        `json:"skipped"`
}

type TestFileCleanup struct {
	FailedResources []TestFailedResource `json:"failed_resources"`
}

type TestFailedResource struct {
	Instance   string `json:"instance"`
	DeposedKey string `json:"deposed_key,omitempty"`
}

type TestFatalInterrupt struct {
	State   []TestFailedResource            `json:"state,omitempty"`
	States  map[string][]TestFailedResource `json:"states,omitempty"`
	Planned []string                        `json:"planned,omitempty"`
}
