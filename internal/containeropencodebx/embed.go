package containeropencodebx

import "embed"

//go:embed spec.yaml.tmpl
var specTemplate string

// globalFS holds the embedded global opencode extension tree.
// Structure under "global/":
//
//	skills/<name>/SKILL.md     → ~/.config/opencode/skills/<name>/SKILL.md
//	agents/<name>.md           → ~/.config/opencode/agents/<name>.md
//	commands/<name>.md         → ~/.config/opencode/commands/<name>.md
//	plugins/<name>.ts|.js      → ~/.config/opencode/plugins/<name>.ts|.js
//
// No opencode.json lives here; configuration is delivered via the
// OPENCODE_CONFIG_CONTENT env var rendered at kit generation time.
//
//go:embed all:global
var globalFS embed.FS
