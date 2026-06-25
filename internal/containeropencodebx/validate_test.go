package containeropencodebx

import (
	"testing"

	E "github.com/IBM/fp-go/v2/either"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeInput_AllDefaults(t *testing.T) {
	input := Input{APIKey: testAPIKey}
	result := NormalizeInput(input)

	assert.Equal(t, DefaultOutputPath, result.OutputPath)
	assert.Equal(t, DefaultAgentImage, result.AgentImage)
	assert.Equal(t, DefaultProvider, result.Provider)
	assert.Equal(t, DefaultGolangciLintVersion, result.GolangciLintVersion)
	assert.Equal(t, DefaultGotestsumVersion, result.GotestsumVersion)
	assert.Equal(t, DefaultMoxideVersion, result.MoxideVersion)
	assert.Equal(t, DefaultSemVersion, result.SemVersion)
	assert.Equal(t, DefaultRtkVersion, result.RtkVersion)
	assert.NotEmpty(t, result.KitName)
}

func TestNormalizeInput_KitNamePattern(t *testing.T) {
	input := Input{APIKey: testAPIKey, KitName: ""}
	result := NormalizeInput(input)
	assert.Regexp(t, `^opencode-sbx[a-zA-Z0-9]{6}$`, result.KitName)
}

func TestValidateInput_MissingAPIKey(t *testing.T) {
	either := ValidateInput(Input{})
	assert.True(t, E.IsLeft(either))
}

func TestValidateInput_InvalidProvider(t *testing.T) {
	input := Input{APIKey: testAPIKey, Provider: "unsupported"}
	either := ValidateInput(input)
	require.True(t, E.IsLeft(either))
	err := E.ToError(either)
	assert.Contains(t, err.Error(), "unsupported provider")
}

func TestValidateInput_ValidProviders(t *testing.T) {
	providers := []string{"openrouter", "anthropic", "openai", "google"}
	for _, p := range providers {
		t.Run(p, func(t *testing.T) {
			input := Input{APIKey: testAPIKey, Provider: p}
			either := ValidateInput(input)
			assert.True(t, E.IsRight(either))
		})
	}
}

func TestValidateInput_OutputParentNotExist(t *testing.T) {
	input := Input{
		APIKey:     testAPIKey,
		OutputPath: "/nonexistent/dir/kit.zip",
	}
	either := ValidateInput(input)
	assert.True(t, E.IsLeft(either))
}

func TestValidateInput_ValidExplicit(t *testing.T) {
	tmpDir := t.TempDir()
	input := Input{
		APIKey:              testAPIKey,
		Provider:            providerAnthropic,
		OutputPath:          tmpDir + "/kit.zip",
		KitName:             testKitName,
		AgentImage:          "custom/image:v1",
		GolangciLintVersion: testGLVersion,
	}
	either := ValidateInput(input)
	assert.True(t, E.IsRight(either))
}
