package internal

import (
	"github.com/hashicorp/terraform/command/cliconfig"
)

// isPluginCacheUsed returns true if the terraform plugin cache dir is specified, https://developer.hashicorp.com/terraform/cli/config/config-file#provider-plugin-cache
func isPluginCacheUsed() bool {
	cfg, _ := cliconfig.LoadConfig()
	return cfg.PluginCacheDir != ""
}
