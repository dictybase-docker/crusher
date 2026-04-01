# Review: Embed Dockerfile Implementations

Comparative review of 4 worktree implementations of `docs/PLAN-embed-dockerfile.md`.
Each worktree was implemented by a separate LLM. All commits after `c94043c` represent
the changes made.

---

## Worktrees

| Branch | Commits after `c94043c` | Description |
|---|---|---|
| `gemini3` | 3 | Logical grouping |
| `glm5` | 1 | Single squashed commit |
| `gpt54` | 2 | Two commits |
| `sonnet4` | 8 | One per plan step |

---

## Baseline

All 4 worktrees: **build succeeds**, **15/15 tests pass**, **0 lint issues**.
All touch the same 11 files. The plan was followed faithfully across the board.
Differences are in craftsmanship details.

---

## Full Scorecard

| Dimension | Sonnet 4 | Gemini 3 | GLM 5 | GPT 4.5 |
|---|:---:|:---:|:---:|:---:|
| **Plan adherence** | Exact | Exact | Exact | Exact |
| **Correctness** | Full | Full | Full | **Bug** |
| **`gosec` nolint** | Inline `//nolint:gosec` | Above-line `//nolint:gosec` with comment | Function-level `//nolint:gosec` with G204 | **Missing** (avoided by using `containerBinary` constant) |
| **Doc comments** | Full (matches plan verbatim) | Full (matches plan, removes pkg docs) | Abbreviated (shorter EmbeddedResolver doc) | **None** on exported types/functions |
| **Package doc preserved** | Yes (both `args.go` and `command.go`) | **No** (removed from both) | Yes (both) | **No** (removed from both) |
| **`exec.go` uses `spec.Bin`** | Yes | Yes | Yes | **No** — hardcodes `containerBinary` |
| **IOEither type param style** | `IOE.FromEither[error, DockerfileResource]` | `IOE.FromEither[error, DockerfileResource]` | `IOE.FromEither[error]` (inferred) | `IOE.FromEither[error]` (inferred) |
| **Lines changed** | 349+/97- | 350+/101- | 338+/97- | 282+/106- |

---

## Detailed Findings

### Sonnet 4 (`sonnet4`)

- Exact plan adherence with granular commit history (8 commits matching the plan's
  8-step implementation order).
- Full doc comments on all exported symbols, matching plan text.
- Preserves existing package doc comments on `args.go` and `command.go`.
- `nolint:gosec` placed inline on the exact `CommandContext` line — the most precise
  suppression scope.
- Includes the `WithResource` curried Kleisli explanation in the `Execute` doc comment.

### Gemini 3 (`gemini3`)

- Functionally identical to Sonnet 4's code output.
- **Removes package doc comments** from `args.go` and `command.go` — unnecessary churn.
- `nolint:gosec` on the line above with a descriptive comment ("this is expected to
  execute user-provided commands") — slightly misleading since the binary is always
  `"container"`, not truly user-provided.
- 3 well-structured logical commits.

### GLM 5 (`glm5`)

- Functionally identical code logic to Sonnet 4 and Gemini 3.
- Preserves package doc comments.
- Shorter doc comments on `EmbeddedResolver` — omits the `04_interface_helpers.go`
  pattern reference. Not a problem, but less helpful for future maintainers.
- `nolint:gosec` placed as function-level directive with G204 code — broadest
  suppression scope. Works, but suppresses all gosec findings in `runProcess`, not
  just the one line.
- Single squashed commit.
- Uses `IOE.FromEither[error]` with inferred type param in `FileResolver` — works
  because Go infers `DockerfileResource`, but is less explicit than the plan's
  `IOE.FromEither[error, DockerfileResource]`.

### GPT 4.5 (`gpt54`) — Has a correctness issue

- **Bug in `exec.go:32`**: `runProcess` uses `exec.CommandContext(ctx, containerBinary)`
  instead of `exec.CommandContext(ctx, spec.Bin)`. This hardcodes the binary name and
  ignores `spec.Bin`. Currently works because `RenderCommand` always sets
  `Bin: containerBinary`, but it breaks the `CommandSpec` abstraction — if `Bin` ever
  changes, `runProcess` silently ignores it. The plan explicitly shows `spec.Bin`.
- **No `gosec` nolint directive** — sidestepped the linter issue by using the constant
  directly, masking the design intent.
- **No doc comments** on any exported type, function, or variable (`DockerfileResource`,
  `FileResolver`, `EmbeddedResolver`, `releaseResource`, `useResource`, `Execute`,
  `runProcess`, `InputFromCommand`).
- Removes package doc comments from `args.go` and `command.go`.
- Smallest diff (282 lines) — but entirely because comments were stripped.
- Removes var comments (`isBlank`, `isAllNonBlank`) from `validate.go` that were
  present in the original code.

---

## Sonnet 4 vs GLM 5 — Focused Comparison

The two strongest implementations. Functional code, types, pipelines, and tests are
identical. Differences are purely in documentation and lint suppression style:

| Dimension | Sonnet 4 | GLM 5 |
|---|---|---|
| **`exec.go` — `nolint:gosec` placement** | Inline on the exact line: `spec.Bin) //nolint:gosec` | Function-level directive above `runProcess` — suppresses all gosec findings in the function, not just one line |
| **`exec.go` — `Execute` doc comment** | Includes extra paragraph explaining `WithResource` curried Kleisli signature | Omits the Kleisli signature explanation |
| **`resource.go` — `EmbeddedResolver` doc** | Includes `04_interface_helpers.go` pattern reference (10 extra lines of godoc) | Omits the pattern reference — shorter doc |
| **`resource.go` — `IOE.FromEither` type params** | Explicit: `IOE.FromEither[error, DockerfileResource]` | Inferred: `IOE.FromEither[error]` (compiler infers second param) |
| **Package doc comments** | Preserved on `args.go` and `command.go` | Preserved on `args.go` and `command.go` |
| **Var/predicate comments in `validate.go`** | Preserved (`isBlank`, `isAllNonBlank` comments kept) | Preserved |
| **Lines changed** | 349+ / 97- | 338+ / 97- |
| **Line delta source** | ~11 extra lines are solely doc comments in `resource.go` and `exec.go` | Leaner — all functional code identical |
| **Functional code** | Identical | Identical |
| **Tests** | Identical (same `resource_test.go`, `validate_test.go`, `args_test.go`) | Identical |

---

## Recommendation

**Sonnet 4** or **GLM 5** — either is production-ready. Pick based on preference:

- **Sonnet 4** if you value thorough doc comments and precise `nolint` scope.
- **GLM 5** if you prefer leaner godoc and are comfortable with function-level lint
  suppression.

**Avoid GPT 4.5** — the `spec.Bin` bug breaks the `CommandSpec` abstraction and the
complete absence of doc comments deviates from project conventions.

**Gemini 3** is functionally sound but unnecessarily removes package doc comments.
