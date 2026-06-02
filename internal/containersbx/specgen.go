package containersbx

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	F "github.com/IBM/fp-go/v2/function"
	IOE "github.com/IBM/fp-go/v2/ioeither"
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
	ConfigDelimiter     string
	SkillsEnvVar        string
}

// GenerateSpec renders the spec template from gs.input and gs.configContent.
func GenerateSpec(gs genState) IOE.IOEither[error, genState] {
	return F.Pipe1(
		IOE.TryCatchError(func() (string, error) {
			data := specTemplateData{
				KitName:             gs.input.KitName,
				GoVersion:           gs.input.GoVersion,
				CrushVersion:        gs.input.CrushVersion,
				GolangciLintVersion: gs.input.GolangciLintVersion,
				GotestsumVersion:    gs.input.GotestsumVersion,
				MoxideVersion:       gs.input.MoxideVersion,
				SemVersion:          gs.input.SemVersion,
				RtkVersion:          gs.input.RtkVersion,
				SkillsEnvVar:        generateSkillsEnvVar(gs.input.SkillsAbsPath),
			}

			data.ConfigContent, data.ConfigDelimiter = escapeForYAMLLiteral(gs.configContent)

			tmpl, err := template.New("spec").Parse(specTemplate)
			if err != nil {
				return "", fmt.Errorf("failed to parse spec template: %w", err)
			}

			var buf bytes.Buffer
			if err := tmpl.Execute(&buf, data); err != nil {
				return "", fmt.Errorf("failed to render spec: %w", err)
			}

			return buf.String(), nil
		}),
		IOE.Map[error](func(spec string) genState {
			return genState{
				input:         gs.input,
				configContent: gs.configContent,
				spec:          spec,
			}
		}),
	)
}

// escapeForYAMLLiteral ensures content doesn't contain the heredoc delimiter "CRUSHCFG".
// Returns (content, safe delimiter string).
func escapeForYAMLLiteral(content string) (string, string) {
	const delimiter = "CRUSHCFG"
	safe := delimiter
	for strings.Contains(content, safe) {
		safe = incrementDelimiter(safe)
	}
	return content, safe
}

// incrementDelimiter appends/increments a numeric suffix on the delimiter.
func incrementDelimiter(d string) string {
	// Strip existing numeric suffix
	base := strings.TrimRight(d, "0123456789")
	num := strings.TrimPrefix(d, base)
	if num == "" {
		return base + "1"
	}
	// Parse and increment
	var n int
	fmt.Sscanf(num, "%d", &n)
	return fmt.Sprintf("%s%d", base, n+1)
}

// generateSkillsEnvVar generates the CRUSH_SKILLS_DIR env var if a skills mount path is set.
func generateSkillsEnvVar(skillsAbsPath string) string {
	if skillsAbsPath == "" {
		return ""
	}
	return fmt.Sprintf(`    CRUSH_SKILLS_DIR: %q`, skillsAbsPath) + "\n"
}
