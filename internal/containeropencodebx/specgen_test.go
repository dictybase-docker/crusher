package containeropencodebx

import (
	"testing"

	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
			pc := providerConfigs[c.name]
			assert.Equal(t, c.authHeader, pc.AuthHeader)
		})
	}
}

func TestLookupProvider_UnknownProvider(t *testing.T) {
	either := ValidateInput(Input{APIKey: testAPIKey, Provider: "bogus"})
	assert.True(t, E.IsLeft(either))
}

func TestGenerateSpec_ContainsResolvedProviderFields(t *testing.T) {
	input := Input{
		KitName:             testKitName,
		AgentImage:          "custom/image:v1",
		GolangciLintVersion: testGLVersion,
		Provider:            providerAnthropic,
		ResolvedProvider:    providerConfigs[providerAnthropic],
	}
	either := GenerateSpec(input)()
	require.True(t, E.IsRight(either))

	gs := E.Fold(
		func(error) genState { return genState{} },
		F.Identity[genState],
	)(either)

	assert.Contains(t, gs.spec, `image: "custom/image:v1"`)
	assert.Contains(t, gs.spec, domainAnthropic+`: anthropic`)
	assert.Contains(t, gs.spec, `headerName: x-api-key`)
	assert.Contains(t, gs.spec, `- ANTHROPIC_API_KEY`)
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
				ResolvedProvider:    providerConfigs[p],
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
				ResolvedProvider:    providerConfigs[c.provider],
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
		ResolvedProvider:    providerConfigs[providerOpenRouter],
	}
	either := GenerateSpec(input)()
	require.True(t, E.IsRight(either))
	gs := E.Fold(
		func(error) genState { return genState{} },
		F.Identity[genState],
	)(either)
	assert.Contains(t, gs.spec, "OPENCODE_CONFIG_CONTENT:")
	assert.Contains(t, gs.spec, `"autoupdate":false`)
	assert.Contains(t, gs.spec, `"edit":"allow"`)
	assert.Contains(t, gs.spec, `"bash":"allow"`)
}

func TestGenerateSpec_ContainsKitName(t *testing.T) {
	input := Input{
		KitName:             "display-name-test",
		AgentImage:          DefaultAgentImage,
		GolangciLintVersion: DefaultGolangciLintVersion,
		Provider:            providerOpenRouter,
		ResolvedProvider:    providerConfigs[providerOpenRouter],
	}
	either := GenerateSpec(input)()
	require.True(t, E.IsRight(either))
	gs := E.Fold(
		func(error) genState { return genState{} },
		F.Identity[genState],
	)(either)
	assert.Contains(t, gs.spec, "displayName: display-name-test")
}
