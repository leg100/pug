package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_IsPluginCacheUsed_WithConfig(t *testing.T) {
	t.Setenv("TF_CLI_CONFIG_FILE", "./testdata/_terraformrc_with_plugin_cache")
	assert.True(t, isPluginCacheUsed())

	t.Setenv("TF_CLI_CONFIG_FILE", "./testdata/_terraformrc_without_plugin_cache")
	assert.False(t, isPluginCacheUsed())
}

func Test_IsPluginCacheUsed_WithEnvVar(t *testing.T) {
	// override config file location, in case the test is run on a computer with
	// a default config file that enables a plugin cache
	t.Setenv("TF_CLI_CONFIG_FILE", "./testdata/_terraformrc_without_plugin_cache")
	assert.False(t, isPluginCacheUsed())

	t.Setenv("TF_PLUGIN_CACHE_DIR", "./testdata/plugin_cache")
	assert.True(t, isPluginCacheUsed())
}
