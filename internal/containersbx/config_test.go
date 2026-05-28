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
	eitherCfg := ReadConfig("")()
	assert.True(t, E.IsRight(eitherCfg))
}

func TestReadConfig_ValidFile(t *testing.T) {
	content := `{"model": "test-model"}`
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "crush.json")
	require.NoError(t, os.WriteFile(configPath, []byte(content), 0o600))

	eitherCfg := ReadConfig(configPath)()
	assert.True(t, E.IsRight(eitherCfg))
	cfg, _ := E.Unwrap(eitherCfg)
	assert.Equal(t, content, cfg)
}

func TestReadConfig_MissingFile(t *testing.T) {
	eitherCfg := ReadConfig("/nonexistent/crush.json")()
	assert.True(t, E.IsLeft(eitherCfg))
}
