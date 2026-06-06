// Package containersbx implements the "sbx" subcommand, which generates a
// Docker Sandbox agent kit (spec.yaml), validates it, packs it into a zip,
// and optionally creates the sandbox instance.
package containersbx

import (
	"context"
)

const (
	// DefaultOutputPath is the default path for the packed kit zip file.
	DefaultOutputPath = "crush-sbx-kit.zip"

	// DefaultAgentImage is the default base image for the sbx sandbox agent.
	DefaultAgentImage = "ghcr.io/dictybase/crusher:sbx-base"

	// DefaultGoVersion is the default Go toolchain version.
	DefaultGoVersion = "1.26.3"

	// agentKitName is the kit schema name declared in spec.yaml (name: crush).
	// sbx create requires this exact value as its positional argument.
	agentKitName = "crush"

	// DefaultCrushVersion is the default Crush version.
	DefaultCrushVersion = "latest"

	// DefaultGolangciLintVersion is the default golangci-lint version.
	DefaultGolangciLintVersion = "2.11.4"

	// DefaultGotestsumVersion is the default gotestsum version.
	DefaultGotestsumVersion = "latest"

	// DefaultMoxideVersion is the default markdown-oxide version.
	DefaultMoxideVersion = "latest"

	// DefaultSemVersion is the default sem version.
	DefaultSemVersion = "latest"

	// DefaultRtkVersion is the default rtk version.
	DefaultRtkVersion = "latest"

	// sbxBinary is the name of the sbx CLI binary.
	sbxBinary = "sbx"
)

// Input holds all parameters for kit generation and optional sandbox creation.
type Input struct {
	OutputPath          string // Path for the packed kit zip
	ConfigPath          string // Path to user's crush.json (empty = use default)
	SkillsPath          string // Path to skills directory
	SkillsAbsPath       string // Resolved absolute path for read-only workspace mount (set by generateToTempDir)
	KitName             string // Sandbox display name
	APIKey              string // OpenRouter API key (required)
	ShouldCreate        bool   // Whether to also create the sandbox instance
	AgentImage          string // Base Docker image for the sandbox agent
	CrushVersion        string // Crush version
	GolangciLintVersion string // golangci-lint version
	GoVersion           string // Go toolchain version
	GotestsumVersion    string // gotestsum version
	MoxideVersion       string // markdown-oxide version
	SemVersion          string // sem version
	RtkVersion          string // rtk version
	Ctx                 context.Context
}

// ResolvedInput holds all resolved data ready for the external sbx pipeline.
type ResolvedInput struct {
	OutputPath   string // Absolute path for the packed kit zip
	TempDir      string // Temp directory containing spec.yaml
	KitName      string // Sandbox display name
	APIKey       string // OpenRouter API key
	ShouldCreate bool   // Whether to create the sandbox instance
	Ctx          context.Context
}

// CommandSpec holds a resolved executable, argv slice, and optional stdin.
type CommandSpec struct {
	Bin   string   // "sbx"
	Args  []string // Full argument list
	Stdin string   // Optional stdin content (piped to process stdin)
}

// KitResult holds the result of kit generation and optional sandbox creation.
type KitResult struct {
	OutputPath string // Packed kit zip path
	KitName    string // Sandbox display name
	Created    bool   // Whether the sandbox instance was created
}

// execState carries state through the Execute pipeline.
type execState struct {
	Input
	TempDir    string
	OutputPath string
	KitName    string
	APIKey     string
	SbxPath    string // resolved sbx binary path (set by runSbxCommand)
	Result     KitResult
}
