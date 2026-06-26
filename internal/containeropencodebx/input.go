// Package containeropencodebx implements the "opencode-sbx" subcommand, which
// generates a Docker Sandbox kit for opencode (spec.yaml + embedded global
// skills/agents/commands/plugins), validates it, packs it into a zip, and
// optionally creates the sandbox instance. Unlike the crush sbx kit, all tools
// are pre-baked into the base image; the kit is configuration-only and the
// opencode config is delivered via the OPENCODE_CONFIG_CONTENT env var.
package containeropencodebx

import (
	"context"
	"os"

	F "github.com/IBM/fp-go/v2/function"
	IOE "github.com/IBM/fp-go/v2/ioeither"
)

// ── Constants ──────────────────────────────────────────────────────────────

const (
	// DefaultOutputPath is the default path for the packed kit zip file.
	DefaultOutputPath = "opencode-sbx-kit.zip"

	// DefaultAgentImage is the default base image for the opencode sbx agent.
	// It extends docker/sandbox-templates:opencode with pre-baked Go tooling.
	DefaultAgentImage = "ghcr.io/dictybase/crusher:opencode-sbx-base"

	// DefaultGolangciLintVersion is the default golangci-lint version.
	// It is the only version flag retained because it is the sole value that
	// surfaces in the generated spec.yaml agentContext text. The gotestsum,
	// moxide, sem, and rtk version flags were removed because the packed kit
	// does not vary with them — the sandbox image ships fixed versions.
	DefaultGolangciLintVersion = "2.11.4"

	// DefaultProvider is the default AI provider.
	DefaultProvider = providerOpenRouter

	// agentKitName is the kit schema name declared in spec.yaml (name: opencode).
	// sbx create requires this exact value as its positional argument.
	agentKitName = "opencode"

	// sbxBinary is the name of the sbx CLI binary.
	sbxBinary = "sbx"

	// createCmd is the name of the sbx create subcommand.
	createCmd = "create"

	// kitNamePrefix is prepended to the random suffix when auto-generating a kit name.
	kitNamePrefix = "opencode-sbx"

	// charNo is the length of the random suffix appended to auto-generated kit names.
	charNo = 6

	// filePerm is the permission bits for written kit files.
	filePerm os.FileMode = 0o644

	// dirPerm is the permission bits for created kit directories.
	dirPerm os.FileMode = 0o755
)

// Provider identifier constants (also used as providerConfigs map keys and in
// tests), kept as named constants to satisfy goconst.
const (
	providerOpenRouter = "openrouter"
	providerAnthropic  = "anthropic"
	providerOpenAI     = "openai"
	providerGoogle     = "google"

	// Repeated network/header literals used by providerConfigs and tests.
	domainOpenRouter = "openrouter.ai"
	domainAnthropic  = "api.anthropic.com"
	domainOpenAI     = "api.openai.com"
	domainGoogle     = "generativelanguage.googleapis.com"

	headerAuthorization = "Authorization"
)

// ── Domain types ───────────────────────────────────────────────────────────

// ProviderConfig holds all template-rendering values for one AI provider.
type ProviderConfig struct {
	ServiceDomain   string // exact hostname for network.serviceDomains
	ProviderDomain  string // same or broader, for network.allowedDomains
	AuthHeader      string // HTTP header name injected by the sbx proxy
	AuthValueFormat string // format string; %s is replaced with the credential
	APIKeyEnvVar    string // env var name on the host and inside the sandbox
}

// providerConfigs is the registry of all supported providers.
// Validated by ValidateInput via R.Lookup.
//
//nolint:gosec // env var names, not hardcoded credentials
var providerConfigs = map[string]ProviderConfig{
	providerOpenRouter: {
		ServiceDomain:   domainOpenRouter,
		ProviderDomain:  domainOpenRouter,
		AuthHeader:      headerAuthorization,
		AuthValueFormat: "Bearer %s",
		APIKeyEnvVar:    "OPENROUTER_API_KEY",
	},
	providerAnthropic: {
		ServiceDomain:   domainAnthropic,
		ProviderDomain:  domainAnthropic,
		AuthHeader:      "x-api-key",
		AuthValueFormat: "%s",
		APIKeyEnvVar:    "ANTHROPIC_API_KEY",
	},
	providerOpenAI: {
		ServiceDomain:   domainOpenAI,
		ProviderDomain:  domainOpenAI,
		AuthHeader:      headerAuthorization,
		AuthValueFormat: "Bearer %s",
		APIKeyEnvVar:    "OPENAI_API_KEY",
	},
	providerGoogle: {
		ServiceDomain:   domainGoogle,
		ProviderDomain:  domainGoogle,
		AuthHeader:      "x-goog-api-key",
		AuthValueFormat: "%s",
		APIKeyEnvVar:    "GOOGLE_API_KEY",
	},
}

// Input holds the raw CLI arguments before normalization or validation.
// The ResolvedProvider field is populated by validateProvider so downstream
// pipeline steps (spec generation) never need to re-lookup the provider.
type Input struct {
	OutputPath          string
	KitName             string
	APIKey              string
	Provider            string         // "openrouter" | "anthropic" | "openai" | "google"
	ResolvedProvider    ProviderConfig // set by validateProvider
	ShouldCreate        bool
	AgentImage          string
	GolangciLintVersion string
	Ctx                 context.Context
}

// KitResult is the final output surfaced to the user.
type KitResult struct {
	OutputPath string
	KitName    string
	Created    bool
}

// CommandSpec holds a resolved sbx CLI invocation.
type CommandSpec struct {
	Ctx   context.Context
	Bin   string
	Args  []string
	Stdin string
}

// genState threads intermediate values through the generateToTempDir pipeline.
// It has no configContent field because opencode config is an env var, not a file.
type genState struct {
	input   Input
	spec    string // rendered spec.yaml content
	tempDir string // absolute path to temp kit directory (set by makeTempDir)
}

// execState carries resolved values through the Execute pipeline.
type execState struct {
	Input
	TempDir    string
	OutputPath string // absolute path for the output .zip
	KitName    string
	APIKey     string
	Result     KitResult
}

// processRunner is the type of the sbx subprocess executor.
// Defined as a named type to enable injection of test doubles.
type processRunner func(CommandSpec) IOE.IOEither[error, F.Void]

// stepState pairs execState with its runner, making every pipeline step a
// univariate function stepState → IOE[stepState].
type stepState struct {
	State execState
	Run   processRunner
}
