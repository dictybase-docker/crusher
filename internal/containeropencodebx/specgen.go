package containeropencodebx

import (
	"bytes"
	"fmt"
	"text/template"

	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	J "github.com/IBM/fp-go/v2/json"
	R "github.com/IBM/fp-go/v2/record"
)

// openCodeConfig is the JSON payload injected via OPENCODE_CONFIG_CONTENT.
type openCodeConfig struct {
	Autoupdate bool           `json:"autoupdate"`
	Permission map[string]any `json:"permission"`
}

// specTemplateData holds the data used to render spec.yaml.tmpl.
type specTemplateData struct {
	KitName             string
	AgentImage          string
	GolangciLintVersion string
	// Resolved from providerConfigs[input.Provider]:
	ProviderID      string // the map key: "openrouter", "anthropic", etc.
	ServiceDomain   string
	ProviderDomain  string
	AuthHeader      string
	AuthValueFormat string
	APIKeyEnvVar    string
	// Compact JSON rendered into OPENCODE_CONFIG_CONTENT env var:
	OpenCodeConfigContent string
}

// buildOpenCodeConfigContent marshals the sandbox-default opencode config to
// compact JSON. Pure — no side-effects — so it stays in the Either layer.
func buildOpenCodeConfigContent() E.Either[error, string] {
	return F.Pipe1(
		J.Marshal(openCodeConfig{
			Autoupdate: false,
			Permission: map[string]any{
				"edit": "allow",
				"bash": "allow",
			},
		}),
		E.Map[error](func(b []byte) string { return string(b) }),
	)
}

// lookupProvider resolves a provider key against providerConfigs. Pure lookup.
func lookupProvider(provider string) E.Either[error, ProviderConfig] {
	return F.Pipe1(
		R.Lookup[ProviderConfig](provider)(providerConfigs),
		E.FromOption[ProviderConfig](func() error {
			return fmt.Errorf("unsupported provider %q", provider)
		}),
	)
}

// withSpecProvider carries the resolved ProviderConfig through the staged
// Either context assembled by E.Do/E.Bind.
type withSpecProvider struct {
	ProviderConfig
}

// withSpecConfig adds the rendered OPENCODE_CONFIG_CONTENT payload.
type withSpecConfig struct {
	withSpecProvider
	OpenCodeConfigContent string
}

var setSpecProvider = F.Curry2(
	func(pc ProviderConfig, _ struct{}) withSpecProvider {
		return withSpecProvider{ProviderConfig: pc}
	},
)

var setSpecConfig = F.Curry2(
	func(content string, ctx withSpecProvider) withSpecConfig {
		return withSpecConfig{
			withSpecProvider:      ctx,
			OpenCodeConfigContent: content,
		}
	},
)

// buildSpecData assembles the template data from input. Both lookupProvider
// and buildOpenCodeConfigContent are pure Either operations, composed here with
// E.Do/E.Bind and lifted to IOEither at the boundary.
func buildSpecData(input Input) IOE.IOEither[error, specTemplateData] {
	return F.Pipe4(
		E.Do[error](struct{}{}),
		E.Bind(
			setSpecProvider,
			func(struct{}) E.Either[error, ProviderConfig] {
				return lookupProvider(input.Provider)
			},
		),
		E.Bind(
			setSpecConfig,
			func(withSpecProvider) E.Either[error, string] {
				return buildOpenCodeConfigContent()
			},
		),
		E.Map[error](func(ctx withSpecConfig) specTemplateData {
			return specTemplateData{
				KitName:               input.KitName,
				AgentImage:            input.AgentImage,
				GolangciLintVersion:   input.GolangciLintVersion,
				ProviderID:            input.Provider,
				ServiceDomain:         ctx.ServiceDomain,
				ProviderDomain:        ctx.ProviderDomain,
				AuthHeader:            ctx.AuthHeader,
				AuthValueFormat:       ctx.AuthValueFormat,
				APIKeyEnvVar:          ctx.APIKeyEnvVar,
				OpenCodeConfigContent: ctx.OpenCodeConfigContent,
			}
		}),
		IOE.FromEither[error],
	)
}

// parseAndRenderTemplate is a Kleisli arrow: specTemplateData → IOEither[error, string].
func parseAndRenderTemplate(data specTemplateData) IOE.IOEither[error, string] {
	return F.Pipe2(
		IOE.TryCatchError(func() (*template.Template, error) {
			return template.New("spec").Parse(specTemplate)
		}),
		IOE.Chain(func(tmpl *template.Template) IOE.IOEither[error, bytes.Buffer] {
			return IOE.TryCatchError(func() (bytes.Buffer, error) {
				var buf bytes.Buffer
				return buf, tmpl.Execute(&buf, data)
			})
		}),
		IOE.Map[error](func(buf bytes.Buffer) string { return buf.String() }),
	)
}

// GenerateSpec renders the spec template from input. Takes Input directly
// (not genState) because there is no configContent to thread — the config is
// produced internally by buildSpecData.
func GenerateSpec(input Input) IOE.IOEither[error, genState] {
	return F.Pipe2(
		buildSpecData(input),
		IOE.Chain(parseAndRenderTemplate),
		IOE.Map[error](func(spec string) genState {
			return genState{input: input, spec: spec}
		}),
	)
}
