module github.com/leg100/pug

go 1.22

require (
	github.com/awalterschulze/gographviz v2.0.3+incompatible
	github.com/charmbracelet/bubbles v0.20.0
	github.com/charmbracelet/bubbletea v1.1.0
	github.com/charmbracelet/lipgloss v0.13.0
	github.com/charmbracelet/x/exp/teatest v0.0.0-20240329185201-62a6965a9fad
	github.com/davecgh/go-spew v1.1.1
	github.com/go-logfmt/logfmt v0.6.0
	github.com/google/uuid v1.6.0
	github.com/hashicorp/hcl/v2 v2.20.1
	github.com/hokaccha/go-prettyjson v0.0.0-20211117102719-0474bc63780f
	github.com/leg100/go-runewidth v0.0.16-0.20240513191656-9e28d2bebd46
	github.com/leg100/reflow v0.0.0-20240513191534-e77d7e432a72
	github.com/mitchellh/iochan v1.0.0
	github.com/muesli/ansi v0.0.0-20230316100256-276c6243b2f6
	github.com/otiai10/copy v1.14.0
	github.com/peterbourgon/ff/v4 v4.0.0-alpha.4
	github.com/stretchr/testify v1.9.0
	golang.org/x/exp v0.0.0-20240506185415-9bf2ced13842
	gopkg.in/yaml.v3 v3.0.1
)

// Many functions of terraform was converted to internal to avoid use as a
// library after v0.15.3. This means that we can't use terraform as a library
// after v0.15.3, so we pull that in here.
require github.com/hashicorp/terraform v0.15.3

require (
	github.com/agext/levenshtein v1.2.3 // indirect
	github.com/apparentlymart/go-textseg/v15 v15.0.0 // indirect
	github.com/apparentlymart/go-versions v1.0.2 // indirect
	github.com/atotto/clipboard v0.1.4 // indirect
	github.com/aymanbagabas/go-osc52/v2 v2.0.1 // indirect
	github.com/aymanbagabas/go-udiff v0.2.0 // indirect
	github.com/charmbracelet/x/ansi v0.3.1 // indirect
	github.com/charmbracelet/x/exp/golden v0.0.0-20240815200342-61de596daa2b // indirect
	github.com/charmbracelet/x/term v0.2.0 // indirect
	github.com/erikgeiser/coninput v0.0.0-20211004153227-1c3628e74d0f // indirect
	github.com/fatih/color v1.17.0 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-hclog v1.6.3 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-retryablehttp v0.7.6 // indirect
	github.com/hashicorp/go-version v1.6.0 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/hashicorp/terraform-svchost v0.1.1 // indirect
	github.com/lucasb-eyer/go-colorful v1.2.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-localereader v0.0.1 // indirect
	github.com/mattn/go-runewidth v0.0.16 // indirect
	github.com/mitchellh/go-wordwrap v1.0.1 // indirect
	github.com/mitchellh/panicwrap v1.0.0 // indirect
	github.com/muesli/cancelreader v0.2.2 // indirect
	github.com/muesli/termenv v0.15.2 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/zclconf/go-cty v1.14.4 // indirect
	golang.org/x/crypto v0.23.0 // indirect
	golang.org/x/mod v0.17.0 // indirect
	golang.org/x/net v0.25.0 // indirect
	golang.org/x/oauth2 v0.20.0 // indirect
	golang.org/x/sync v0.8.0 // indirect
	golang.org/x/sys v0.25.0 // indirect
	golang.org/x/text v0.17.0 // indirect
	golang.org/x/tools v0.21.1-0.20240508182429-e35e4ccd0d2d // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

// NOTE: do not add replace directives; they're incompatible with go install.
// See https://github.com/golang/go/issues/44840
