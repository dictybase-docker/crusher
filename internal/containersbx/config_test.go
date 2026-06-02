package containersbx

import (
	"os"
	"path/filepath"
	"testing"

	E "github.com/IBM/fp-go/v2/either"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	assert.Contains(t, cfg, `"openrouter"`)
	assert.Contains(t, cfg, `"openai/gpt-4o"`)
	assert.Contains(t, cfg, `"https://openrouter.ai/api/v1"`)
	assert.NotContains(t, cfg, `"anthropic"`)
	assert.NotContains(t, cfg, `"deepseek"`)
}

func TestReadConfig_BlankPath(t *testing.T) {
	eitherCfg := ReadConfig(Input{})()
	assert.True(t, E.IsRight(eitherCfg))
	gs, _ := E.Unwrap(eitherCfg)
	require.Equal(t, DefaultConfig(), gs.configContent)
}

func TestReadConfig_ValidFile(t *testing.T) {
	content := `{"model": "test-model"}`
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "crush.json")
	require.NoError(t, os.WriteFile(configPath, []byte(content), 0o600))

	eitherCfg := ReadConfig(Input{ConfigPath: configPath})()
	assert.True(t, E.IsRight(eitherCfg))
	gs, _ := E.Unwrap(eitherCfg)
	assert.Equal(t, content, gs.configContent)
}

func TestReadConfig_MissingFile(t *testing.T) {
	eitherCfg := ReadConfig(Input{ConfigPath: "/nonexistent/crush.json"})()
	assert.True(t, E.IsLeft(eitherCfg))
}
