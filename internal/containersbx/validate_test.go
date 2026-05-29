package containersbx

import (
	"os"
	"path/filepath"
	"testing"

	E "github.com/IBM/fp-go/v2/either"
	"github.com/stretchr/testify/assert"
)

func TestNormalizeInput_AllDefaults(t *testing.T) {
	input := Input{APIKey: "test-key"}
	result := NormalizeInput(input)

	assert.NotEmpty(t, result.OutputPath)
	assert.NotEmpty(t, result.KitName)
	assert.NotEmpty(t, result.GoVersion)
	assert.NotEmpty(t, result.CrushVersion)
	assert.NotEmpty(t, result.GolangciLintVersion)
}

func TestNormalizeInput_BlankOutputPath(t *testing.T) {
	input := Input{APIKey: "test-key", OutputPath: ""}
	result := NormalizeInput(input)
	assert.Equal(t, DefaultOutputPath, result.OutputPath)
}

func TestNormalizeInput_BlankKitName(t *testing.T) {
	input := Input{APIKey: "test-key", KitName: ""}
	result := NormalizeInput(input)
	assert.Regexp(t, `^crush-sbx[a-zA-Z0-9]{6}$`, result.KitName)
}

func TestValidateInput_MissingAPIKey(t *testing.T) {
	input := Input{}
	either := ValidateInput(input)
	assert.True(t, E.IsLeft(either))
}

func TestValidateInput_NonExistentConfig(t *testing.T) {
	input := Input{
		APIKey:     "test-key",
		ConfigPath: "/nonexistent/config.json",
	}
	either := ValidateInput(input)
	assert.True(t, E.IsLeft(either))
}

func TestValidateInput_NonExistentSkills(t *testing.T) {
	input := Input{
		APIKey:     "test-key",
		SkillsPath: "/nonexistent/skills",
	}
	either := ValidateInput(input)
	assert.True(t, E.IsLeft(either))
}

func TestValidateInput_AllExplicit(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")
	os.WriteFile(configPath, []byte("{}"), 0o600)

	input := Input{
		APIKey:              "test-key",
		OutputPath:          "test-output.zip",
		ConfigPath:          configPath,
		KitName:             "my-sandbox",
		CrushVersion:        "v1.0.0",
		GolangciLintVersion: "2.0.0",
		GoVersion:           "1.22.0",
	}
	either := ValidateInput(input)
	assert.True(t, E.IsRight(either))
}

func TestValidateInput_OutputParentNotExist(t *testing.T) {
	input := Input{
		APIKey:     "test-key",
		OutputPath: "/nonexistent/dir/file.zip",
	}
	either := ValidateInput(input)
	assert.True(t, E.IsLeft(either))
}

func TestValidateInput_ValidWithSkillsPath(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")
	os.WriteFile(configPath, []byte("{}"), 0o600)

	input := Input{
		APIKey:     "test-key",
		ConfigPath: configPath,
		SkillsPath: tmpDir,
	}
	either := ValidateInput(input)
	assert.True(t, E.IsRight(either))
}

func TestValidateInput_SkillsPathNotDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "not-a-dir")
	os.WriteFile(filePath, []byte("data"), 0o600)

	input := Input{
		APIKey:     "test-key",
		SkillsPath: filePath,
	}
	either := ValidateInput(input)
	assert.True(t, E.IsLeft(either))
	err := E.ToError(either)
	assert.Contains(t, err.Error(), "skills path is not a directory")
}

func TestValidateInput_SkillsPathEmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	input := Input{
		APIKey:     "test-key",
		SkillsPath: tmpDir,
	}
	either := ValidateInput(input)
	assert.True(t, E.IsLeft(either))
	err := E.ToError(either)
	assert.Contains(t, err.Error(), "skills directory is empty")
}

func TestValidateInput_ValidWithConfigPath(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")
	os.WriteFile(configPath, []byte("{}"), 0o600)

	input := Input{
		APIKey:     "test-key",
		ConfigPath: configPath,
	}
	either := ValidateInput(input)
	assert.True(t, E.IsRight(either))
}
