module github.com/leg100/pug

go 1.21.4

require (
	github.com/btcsuite/btcutil v1.0.2
	github.com/google/uuid v1.5.0
	github.com/hashicorp/hcl/v2 v2.19.1

	// Many functions of terraform was converted to internal to avoid use as a library after v0.15.3. This means that we
	// can't use terraform as a library after v0.15.3, so we pull that in here.
	github.com/hashicorp/terraform v0.15.3
	golang.org/x/sync v0.6.0
)

require (
	github.com/charmbracelet/bubbles v0.17.1
	github.com/charmbracelet/bubbletea v0.25.0
	github.com/charmbracelet/lipgloss v0.9.1
	github.com/mitchellh/iochan v1.0.0
	github.com/muesli/reflow v0.3.0
	github.com/peterbourgon/ff/v4 v4.0.0-alpha.4
	github.com/stretchr/testify v1.8.4
	golang.org/x/exp v0.0.0-20240119083558-1b970713d09a
)

require (
	github.com/agext/levenshtein v1.2.2 // indirect
	github.com/apparentlymart/go-textseg/v13 v13.0.0 // indirect
	github.com/apparentlymart/go-textseg/v15 v15.0.0 // indirect
	github.com/apparentlymart/go-versions v1.0.1 // indirect
	github.com/atotto/clipboard v0.1.4 // indirect
	github.com/aymanbagabas/go-osc52/v2 v2.0.1 // indirect
	github.com/containerd/console v1.0.4-0.20230313162750-1ae8d489ac81 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/fatih/color v1.7.0 // indirect
	github.com/golang/protobuf v1.4.3 // indirect
	github.com/google/go-cmp v0.5.8 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.1 // indirect
	github.com/hashicorp/go-hclog v0.15.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-retryablehttp v0.5.2 // indirect
	github.com/hashicorp/go-version v1.2.1 // indirect
	github.com/hashicorp/hcl v0.0.0-20170504190234-a4b07c25de5f // indirect
	github.com/hashicorp/terraform-svchost v0.0.0-20200729002733-f050f53b9734 // indirect
	github.com/lmittmann/tint v1.0.4 // indirect
	github.com/lucasb-eyer/go-colorful v1.2.0 // indirect
	github.com/mattn/go-colorable v0.1.4 // indirect
	github.com/mattn/go-isatty v0.0.18 // indirect
	github.com/mattn/go-localereader v0.0.1 // indirect
	github.com/mattn/go-runewidth v0.0.15 // indirect
	github.com/mitchellh/go-wordwrap v1.0.0 // indirect
	github.com/mitchellh/panicwrap v1.0.0 // indirect
	github.com/muesli/ansi v0.0.0-20211018074035-2e021307bc4b // indirect
	github.com/muesli/cancelreader v0.2.2 // indirect
	github.com/muesli/termenv v0.15.2 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/sahilm/fuzzy v0.1.1-0.20230530133925-c48e322e2a8f // indirect
	github.com/zclconf/go-cty v1.13.0 // indirect
	golang.org/x/crypto v0.0.0-20220517005047-85d78b3ac167 // indirect
	golang.org/x/mod v0.14.0 // indirect
	golang.org/x/net v0.0.0-20211112202133-69e39bad7dc2 // indirect
	golang.org/x/oauth2 v0.0.0-20200902213428-5d25da1a8d43 // indirect
	golang.org/x/sys v0.12.0 // indirect
	golang.org/x/term v0.6.0 // indirect
	golang.org/x/text v0.11.0 // indirect
	google.golang.org/appengine v1.6.6 // indirect
	google.golang.org/protobuf v1.25.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
