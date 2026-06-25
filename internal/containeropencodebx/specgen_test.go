package containeropencodebx

import (
	"encoding/json"
	"testing"

	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildOpenCodeConfigContent_Valid(t *testing.T) {
	either := buildOpenCodeConfigContent()
	require.True(t, E.IsRight(either))

	content := E.Fold(
		func(error) string { return "" },
		F.Identity[string],
	)(either)

	var cfg openCodeConfig
	require.NoError(t, json.Unmarshal([]byte(content), &cfg))
	assert.False(t, cfg.Autoupdate)
	assert.Equal(t, "allow", cfg.Permission["edit"])
	assert.Equal(t, "allow", cfg.Permission["bash"])
}

func TestLookupProvider_KnownProvider(t *testing.T) {
	cases := []struct {
		name       string
		authHeader string
	}{
		{providerOpenRouter, headerAuthorization},
		{providerAnthropic, "x-api-key"},
		{providerOpenAI, headerAuthorization},
		{providerGoogle, "x-goog-api-key"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			either := lookupProvider(c.name)
			require.True(t, E.IsRight(either))
			pc := E.Fold(
				func(error) ProviderConfig { return ProviderConfig{} },
				F.Identity[ProviderConfig],
			)(either)
			assert.Equal(t, c.authHeader, pc.AuthHeader)
		})
	}
}

func TestLookupProvider_UnknownProvider(t *testing.T) {
	either := lookupProvider("bogus")
	assert.True(t, E.IsLeft(either))
}

func TestBuildSpecData_FieldMapping(t *testing.T) {
	input := Input{
		KitName:             testKitName,
		AgentImage:          "custom/image:v1",
		GolangciLintVersion: testGLVersion,
		Provider:            providerAnthropic,
	}
	either := buildSpecData(input)()
	require.True(t, E.IsRight(either))

	data := E.Fold(
		func(error) specTemplateData { return specTemplateData{} },
		F.Identity[specTemplateData],
	)(either)

	require.Equal(t, testKitName, data.KitName)
	require.Equal(t, "custom/image:v1", data.AgentImage)
	require.Equal(t, providerAnthropic, data.ProviderID)
	require.Equal(t, domainAnthropic, data.ServiceDomain)
	require.Equal(t, "x-api-key", data.AuthHeader)
	require.Equal(t, "ANTHROPIC_API_KEY", data.APIKeyEnvVar)

	var cfg openCodeConfig
	require.NoError(t, json.Unmarshal([]byte(data.OpenCodeConfigContent), &cfg))
	assert.False(t, cfg.Autoupdate)
}

func TestGenerateSpec_NoUnresolvedVars(t *testing.T) {
	providers := []string{providerOpenRouter, providerAnthropic, providerOpenAI, providerGoogle}
	for _, p := range providers {
		t.Run(p, func(t *testing.T) {
			input := Input{
				KitName:             testKitName,
				AgentImage:          DefaultAgentImage,
				GolangciLintVersion: DefaultGolangciLintVersion,
				Provider:            p,
			}
			either := GenerateSpec(input)()
			require.True(t, E.IsRight(either))
			gs := E.Fold(
				func(error) genState { return genState{} },
				F.Identity[genState],
			)(either)
			assert.NotContains(t, gs.spec, "{{.")
		})
	}
}

func TestGenerateSpec_ContainsProviderDomain(t *testing.T) {
	cases := []struct {
		provider string
		domain   string
	}{
		{providerOpenRouter, domainOpenRouter},
		{providerAnthropic, domainAnthropic},
		{providerOpenAI, domainOpenAI},
		{providerGoogle, domainGoogle},
	}
	for _, c := range cases {
		t.Run(c.provider, func(t *testing.T) {
			input := Input{
				KitName:             testKitName,
				AgentImage:          DefaultAgentImage,
				GolangciLintVersion: DefaultGolangciLintVersion,
				Provider:            c.provider,
			}
			either := GenerateSpec(input)()
			require.True(t, E.IsRight(either))
			gs := E.Fold(
				func(error) genState { return genState{} },
				F.Identity[genState],
			)(either)
			assert.Contains(t, gs.spec, `"`+c.domain+`:443"`)
		})
	}
}

func TestGenerateSpec_ContainsOpenCodeConfigContent(t *testing.T) {
	input := Input{
		KitName:             testKitName,
		AgentImage:          DefaultAgentImage,
		GolangciLintVersion: DefaultGolangciLintVersion,
		Provider:            providerOpenRouter,
	}
	either := GenerateSpec(input)()
	require.True(t, E.IsRight(either))
	gs := E.Fold(
		func(error) genState { return genState{} },
		F.Identity[genState],
	)(either)
	assert.Contains(t, gs.spec, "OPENCODE_CONFIG_CONTENT:")
	assert.Contains(t, gs.spec, `"autoupdate":false`)
}

func TestGenerateSpec_ContainsKitName(t *testing.T) {
	input := Input{
		KitName:             "display-name-test",
		AgentImage:          DefaultAgentImage,
		GolangciLintVersion: DefaultGolangciLintVersion,
		Provider:            providerOpenRouter,
	}
	either := GenerateSpec(input)()
	require.True(t, E.IsRight(either))
	gs := E.Fold(
		func(error) genState { return genState{} },
		F.Identity[genState],
	)(either)
	assert.Contains(t, gs.spec, "displayName: display-name-test")
}
