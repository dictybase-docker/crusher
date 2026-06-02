package containersbx

import (
	"bytes"
	"fmt"
	"text/template"

	F "github.com/IBM/fp-go/v2/function"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	O "github.com/IBM/fp-go/v2/option"
	Str "github.com/IBM/fp-go/v2/string"
)

// specTemplateData holds the data used to render spec.yaml.tmpl.
type specTemplateData struct {
	KitName             string
	GoVersion           string
	CrushVersion        string
	GolangciLintVersion string
	GotestsumVersion    string
	MoxideVersion       string
	SemVersion          string
	RtkVersion          string
	ConfigContent       string
	SkillsEnvVar        string
}

// GenerateSpec renders the spec template from gs.input and gs.configContent.
func GenerateSpec(gs genState) IOE.IOEither[error, genState] {
	return F.Pipe4(
		gs,
		buildSpecData,
		IOE.Of[error],
		IOE.Chain(parseAndRenderTemplate),
		IOE.Map[error](func(spec string) genState {
			return genState{
				input:         gs.input,
				configContent: gs.configContent,
				spec:          spec,
			}
		}),
	)
}

// buildSpecData constructs the template data from genState. Pure — no I/O.
func buildSpecData(gs genState) specTemplateData {
	return specTemplateData{
		KitName:             gs.input.KitName,
		GoVersion:           gs.input.GoVersion,
		CrushVersion:        gs.input.CrushVersion,
		GolangciLintVersion: gs.input.GolangciLintVersion,
		GotestsumVersion:    gs.input.GotestsumVersion,
		MoxideVersion:       gs.input.MoxideVersion,
		SemVersion:          gs.input.SemVersion,
		RtkVersion:          gs.input.RtkVersion,
		SkillsEnvVar:        generateSkillsEnvVar(gs.input.SkillsAbsPath),
		ConfigContent:       gs.configContent,
	}
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
				err := tmpl.Execute(&buf, data)
				return buf, err
			})
		}),
		IOE.Map[error](func(buf bytes.Buffer) string {
			return buf.String()
		}),
	)
}

// generateSkillsEnvVar generates the CRUSH_SKILLS_DIR env var if a skills mount path is set.
func generateSkillsEnvVar(skillsAbsPath string) string {
	return F.Pipe2(
		skillsAbsPath,
		O.FromPredicate(Str.IsNonEmpty),
		O.Fold(
			F.Constant(""),
			func(p string) string {
				return fmt.Sprintf(
					`    CRUSH_SKILLS_DIR: %q`,
					p,
				) + "\n"
			},
		),
	)
}
