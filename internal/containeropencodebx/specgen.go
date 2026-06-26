package containeropencodebx

import (
	"bytes"
	"text/template"

	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	J "github.com/IBM/fp-go/v2/json"
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
		IOE.Map[error](func(buf bytes.Buffer) string {
			return buf.String()
		}),
	)
}

// GenerateSpec renders the spec template from input. The ProviderConfig
// was already resolved during ValidateInput and is carried on
// input.ResolvedProvider, so the template data can be assembled inline.
func GenerateSpec(input Input) IOE.IOEither[error, genState] {
	pc := input.ResolvedProvider

	return F.Pipe4(
		J.Marshal(openCodeConfig{
			Autoupdate: false,
			Permission: map[string]any{
				"edit": "allow",
				"bash": "allow",
			},
		}),
		E.Map[error](func(b []byte) specTemplateData {
			return specTemplateData{
				KitName:               input.KitName,
				AgentImage:            input.AgentImage,
				GolangciLintVersion:   input.GolangciLintVersion,
				ProviderID:            input.Provider,
				ServiceDomain:         pc.ServiceDomain,
				ProviderDomain:        pc.ProviderDomain,
				AuthHeader:            pc.AuthHeader,
				AuthValueFormat:       pc.AuthValueFormat,
				APIKeyEnvVar:          pc.APIKeyEnvVar,
				OpenCodeConfigContent: string(b),
			}
		}),
		IOE.FromEither[error],
		IOE.Chain(parseAndRenderTemplate),
		IOE.Map[error](func(spec string) genState {
			return genState{input: input, spec: spec}
		}),
	)
}
