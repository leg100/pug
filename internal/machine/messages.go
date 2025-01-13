package machine

import (
	"time"
)

type Message interface {
	CommonFields() Common
}

type Common struct {
	Type      MessageType `json:"type"`
	Level     string      `json:"@level"`
	Message   string      `json:"@message"`
	Module    string      `json:"@module"`
	TimeStamp time.Time   `json:"@timestamp"`
}

func (m Common) CommonFields() Common {
	return m
}

type VersionMsg struct {
	Common
	Terraform string `json:"terraform"`
	Tofu      string `json:"tofu"`
	UI        string `json:"ui"`
}

type LogMsg struct {
	Common
	KVs map[string]interface{}
}

type DiagnosticsMsg struct {
	Common
	Diagnostic *Diagnostic `json:"diagnostic"`
}

type PlannedChangeMsg struct {
	Common
	Change *ResourceInstanceChange `json:"change"`
}

type ChangeSummaryMsg struct {
	Common
	Changes *ChangeSummary `json:"changes"`
}

type OutputMsg struct {
	Common
	Outputs Outputs `json:"outputs"`
}
