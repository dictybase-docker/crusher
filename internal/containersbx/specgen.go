package containersbx

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
)

// specTemplateData holds the data used to render spec.yaml.tmpl.
type specTemplateData struct {
	KitName              string
	GoVersion            string
	CrushVersion         string
	GolangciLintVersion  string
	GotestsumVersion     string
	MoxideVersion        string
	SemVersion           string
	RtkVersion           string
	ConfigContent        string
	ConfigDelimiter      string
	SkillsInstallSection string
	SkillsMemorySection  string
}

// GenerateSpec renders a spec.yaml string from the given Input.
func GenerateSpec(input Input, configContent string, skills map[string]string) (string, error) {
	data := specTemplateData{
		KitName:              input.KitName,
		GoVersion:            input.GoVersion,
		CrushVersion:         input.CrushVersion,
		GolangciLintVersion:  input.GolangciLintVersion,
		GotestsumVersion:     input.GotestsumVersion,
		MoxideVersion:        input.MoxideVersion,
		SemVersion:           input.SemVersion,
		RtkVersion:           input.RtkVersion,
		SkillsInstallSection: generateSkillsInstallSection(skills),
		SkillsMemorySection:  generateSkillsMemorySection(skills),
	}

	data.ConfigContent, data.ConfigDelimiter = escapeForYAMLLiteral(configContent, "CRUSHCFG")

	tmpl, err := template.New("spec").Parse(specTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse spec template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to render spec: %w", err)
	}

	return buf.String(), nil
}

// escapeForYAMLLiteral ensures content doesn't contain the heredoc delimiter.
// Returns (content, safe delimiter string).
func escapeForYAMLLiteral(content string, delimiter string) (string, string) {
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

// generateSkillsInstallSection renders install commands for skills.
func generateSkillsInstallSection(skills map[string]string) string {
	if len(skills) == 0 {
		return ""
	}

	var parts []string
	for name, content := range skills {
		escaped, delim := escapeForYAMLLiteral(content, "CRUSHSKILL")
		parts = append(parts, fmt.Sprintf(`    # Install skill: %s
    - command: |
        mkdir -p /home/agent/crush/skills/%s &&
        cat > /home/agent/crush/skills/%s/SKILL.md << '%s'
%s
%s
      user: "1000"
      description: "Install %s skill"`, name, name, name, delim, escaped, delim, name))
	}
	return strings.Join(parts, "\n") + "\n"
}

// generateSkillsMemorySection renders the skills table for memory.
func generateSkillsMemorySection(skills map[string]string) string {
	if len(skills) == 0 {
		return ""
	}

	var lines []string
	for name := range skills {
		lines = append(lines, fmt.Sprintf("  | `%s` | Custom skill |", name))
	}
	header := "\n  ## Skills\n\n" +
		"  The following custom skills are available:\n\n" +
		"  | Skill | Purpose |\n  |-------|---------|\n"
	return header + strings.Join(lines, "\n") + "\n"
}
