// Package sbxexec provides the provider-agnostic sbx subprocess runner shared by
// the crush and opencode sandbox kit packages. Only the subprocess boundary
// lives here; package-specific pipeline steps continue to live beside their
// package-local state types.
package sbxexec

import "context"

// CommandSpec holds a resolved sbx CLI invocation.
type CommandSpec struct {
	Ctx   context.Context // context for command cancellation/deadline
	Bin   string          // binary name, e.g. "sbx"
	Args  []string        // full argument list
	Stdin string          // optional content piped to stdin
}
