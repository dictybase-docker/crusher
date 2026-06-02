package containersbx

import (
	"testing"

	E "github.com/IBM/fp-go/v2/either"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateSpec_MinimalInput(t *testing.T) {
	input := Input{
		KitName:             "test-sandbox",
		GoVersion:           DefaultGoVersion,
		CrushVersion:        DefaultCrushVersion,
		GolangciLintVersion: DefaultGolangciLintVersion,
		GotestsumVersion:    DefaultGotestsumVersion,
		MoxideVersion:       DefaultMoxideVersion,
		SemVersion:          DefaultSemVersion,
		RtkVersion:          DefaultRtkVersion,
	}
	configContent := DefaultConfig()
	gs := genState{input: input, configContent: configContent}
	result := GenerateSpec(gs)()
	require.True(t, E.IsRight(result))
	spec, _ := E.Unwrap(result)

	assert.Contains(t, spec.spec, "test-sandbox")
	assert.Contains(t, spec.spec, "openrouter.ai")
	assert.Contains(t, spec.spec, DefaultGoVersion)
	assert.Contains(t, spec.spec, DefaultCrushVersion)
	assert.Contains(t, spec.spec, DefaultGolangciLintVersion)
	assert.Contains(t, spec.spec, "openrouter")
	assert.NotContains(t, spec.spec, "{{.")
}

func TestGenerateSpec_CustomConfig(t *testing.T) {
	input := Input{
		KitName:             "custom-sbx",
		GoVersion:           "1.23.0",
		CrushVersion:        "v2.0.0",
		GolangciLintVersion: "2.0.0",
		GotestsumVersion:    DefaultGotestsumVersion,
		MoxideVersion:       DefaultMoxideVersion,
		SemVersion:          DefaultSemVersion,
		RtkVersion:          DefaultRtkVersion,
	}
	configContent := `{"model":"custom-model"}`
	gs := genState{input: input, configContent: configContent}
	result := GenerateSpec(gs)()
	require.True(t, E.IsRight(result))
	spec, _ := E.Unwrap(result)

	assert.Contains(t, spec.spec, "custom-sbx")
	assert.Contains(t, spec.spec, `"model":"custom-model"`)
	assert.Contains(t, spec.spec, "1.23.0")
	assert.Contains(t, spec.spec, "v2.0.0")
}

func TestGenerateSpec_WithSkills(t *testing.T) {
	input := Input{
		KitName:             "skills-sbx",
		GoVersion:           DefaultGoVersion,
		CrushVersion:        DefaultCrushVersion,
		GolangciLintVersion: DefaultGolangciLintVersion,
		GotestsumVersion:    DefaultGotestsumVersion,
		MoxideVersion:       DefaultMoxideVersion,
		SemVersion:          DefaultSemVersion,
		RtkVersion:          DefaultRtkVersion,
		SkillsAbsPath:       "/home/agent/crush/skills",
	}
	gs := genState{input: input, configContent: DefaultConfig()}
	result := GenerateSpec(gs)()
	require.True(t, E.IsRight(result))
	spec, _ := E.Unwrap(result)

	assert.Contains(t, spec.spec, `CRUSH_SKILLS_DIR: "/home/agent/crush/skills"`)
}

func TestGenerateSpec_EmptySkills(t *testing.T) {
	input := Input{
		KitName:             "no-skills",
		GoVersion:           DefaultGoVersion,
		CrushVersion:        DefaultCrushVersion,
		GolangciLintVersion: DefaultGolangciLintVersion,
		GotestsumVersion:    DefaultGotestsumVersion,
		MoxideVersion:       DefaultMoxideVersion,
		SemVersion:          DefaultSemVersion,
		RtkVersion:          DefaultRtkVersion,
	}
	gs := genState{input: input, configContent: DefaultConfig()}
	result := GenerateSpec(gs)()
	require.True(t, E.IsRight(result))
	spec, _ := E.Unwrap(result)

	assert.NotContains(t, spec.spec, "Install skill:")
	assert.NotContains(t, spec.spec, "## Skills")
	assert.NotContains(t, spec.spec, "CRUSH_SKILLS_DIR")
}

func TestGenerateSpec_AllVersions(t *testing.T) {
	input := Input{
		KitName:             "versions-sbx",
		GoVersion:           "1.22.0",
		CrushVersion:        "v3.0.0",
		GolangciLintVersion: "2.5.0",
		GotestsumVersion:    "v1.0.0",
		MoxideVersion:       "v0.5.0",
		SemVersion:          "v4.0.0",
		RtkVersion:          "v2.0.0",
	}
	gs := genState{input: input, configContent: DefaultConfig()}
	result := GenerateSpec(gs)()
	require.True(t, E.IsRight(result))
	spec, _ := E.Unwrap(result)

	assert.Contains(t, spec.spec, "go1.22.0.linux-amd64")
	assert.Contains(t, spec.spec, "crush@v3.0.0")
	assert.Contains(t, spec.spec, `GOLANGCI_LINT_VERSION="2.5.0"`)
	assert.Contains(t, spec.spec, `gotestsum@v1.0.0`)
	assert.Contains(t, spec.spec, `markdown-oxide" "v0.5.0"`)
	assert.Contains(t, spec.spec, `sem" "v4.0.0"`)
	assert.Contains(t, spec.spec, `rtk" "v2.0.0"`)
}

func TestEscapeForYAMLLiteral_NoMatch(t *testing.T) {
	content := "plain content"
	escaped, delim := escapeForYAMLLiteral(content)
	assert.Equal(t, "plain content", escaped)
	assert.Equal(t, "CRUSHCFG", delim)
}

func TestEscapeForYAMLLiteral_OneMatch(t *testing.T) {
	content := "before CRUSHCFG after"
	escaped, delim := escapeForYAMLLiteral(content)
	assert.Equal(t, "before CRUSHCFG after", escaped)
	assert.Equal(t, "CRUSHCFG1", delim)
}

func TestEscapeForYAMLLiteral_MultipleMatches(t *testing.T) {
	content := "CRUSHCFG in here and CRUSHCFG again CRUSHCFG"
	escaped, delim := escapeForYAMLLiteral(content)
	assert.Equal(t, content, escaped)
	assert.Equal(t, "CRUSHCFG1", delim)
}

func TestGenerateSpec_ConfigContainsCRUSHCFG(t *testing.T) {
	input := Input{
		KitName:             "test-sbx",
		GoVersion:           DefaultGoVersion,
		CrushVersion:        DefaultCrushVersion,
		GolangciLintVersion: DefaultGolangciLintVersion,
		GotestsumVersion:    DefaultGotestsumVersion,
		MoxideVersion:       DefaultMoxideVersion,
		SemVersion:          DefaultSemVersion,
		RtkVersion:          DefaultRtkVersion,
	}
	configContent := `{"key": "CRUSHCFG value"}`
	gs := genState{input: input, configContent: configContent}
	result := GenerateSpec(gs)()
	require.True(t, E.IsRight(result))
	spec, _ := E.Unwrap(result)

	assert.Contains(t, spec.spec, "CRUSHCFG1")
	assert.Contains(t, spec.spec, `{"key": "CRUSHCFG value"}`)
}

func TestIncrementDelimiter(t *testing.T) {
	assert.Equal(t, "CRUSHCFG1", incrementDelimiter("CRUSHCFG"))
	assert.Equal(t, "CRUSHCFG2", incrementDelimiter("CRUSHCFG1"))
	assert.Equal(t, "CRUSHCFG10", incrementDelimiter("CRUSHCFG9"))
	assert.Equal(t, "X1", incrementDelimiter("X"))
	assert.Equal(t, "SKILL2", incrementDelimiter("SKILL1"))
}

func TestGenerateSkillsEnvVar_WithPath(t *testing.T) {
	result := generateSkillsEnvVar("/home/agent/crush/skills")
	assert.Contains(t, result, `CRUSH_SKILLS_DIR: "/home/agent/crush/skills"`)
}

func TestGenerateSkillsEnvVar_EmptyPath(t *testing.T) {
	result := generateSkillsEnvVar("")
	assert.Empty(t, result)
}

func TestBuildSpecData_FieldMapping(t *testing.T) {
	gs := genState{
		input: Input{
			KitName:             "my-kit",
			GoVersion:           "1.23.0",
			CrushVersion:        "v2.0.0",
			GolangciLintVersion: "2.0.0",
			GotestsumVersion:    "v0.6.0",
			MoxideVersion:       "v0.4.0",
			SemVersion:          "v3.0.0",
			RtkVersion:          "v1.5.0",
			SkillsAbsPath:       "/skills",
		},
		configContent: `{"key":"value"}`,
	}
	data := buildSpecData(gs)
	require.Equal(t, "my-kit", data.KitName)
	require.Equal(t, "1.23.0", data.GoVersion)
	require.Equal(t, "v2.0.0", data.CrushVersion)
	require.Equal(t, "2.0.0", data.GolangciLintVersion)
	require.Equal(t, "v0.6.0", data.GotestsumVersion)
	require.Equal(t, "v0.4.0", data.MoxideVersion)
	require.Equal(t, "v3.0.0", data.SemVersion)
	require.Equal(t, "v1.5.0", data.RtkVersion)
	require.Equal(t, `{"key":"value"}`, data.ConfigContent)
	require.Equal(t, "CRUSHCFG", data.ConfigDelimiter)
	require.Contains(t, data.SkillsEnvVar, "/skills")
}

func TestBuildSpecData_DelimiterEscaping(t *testing.T) {
	gs := genState{
		input:         Input{KitName: "k"},
		configContent: `contains CRUSHCFG word`,
	}
	data := buildSpecData(gs)
	require.Equal(t, "CRUSHCFG1", data.ConfigDelimiter)
	require.Equal(t, `contains CRUSHCFG word`, data.ConfigContent)
}

func TestParseAndRenderTemplate_Success(t *testing.T) {
	gs := genState{
		input: Input{
			KitName:             "tmpl-test",
			GoVersion:           DefaultGoVersion,
			CrushVersion:        DefaultCrushVersion,
			GolangciLintVersion: DefaultGolangciLintVersion,
			GotestsumVersion:    DefaultGotestsumVersion,
			MoxideVersion:       DefaultMoxideVersion,
			SemVersion:          DefaultSemVersion,
			RtkVersion:          DefaultRtkVersion,
		},
		configContent: DefaultConfig(),
	}
	data := buildSpecData(gs)
	result := parseAndRenderTemplate(data)()
	require.True(t, E.IsRight(result))
	spec, _ := E.Unwrap(result)
	require.Contains(t, spec, "tmpl-test")
	require.NotContains(t, spec, "{{.")
}
