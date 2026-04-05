# Prompt: Implement Dynamic Tool Versioning for `container-cli`

**Objective:**  
Extend the `container-cli build` subcommand to support dynamic versioning for tools installed in the Docker image. This involves adding CLI flags that map to `--build-arg` parameters passed to the underlying `container build` command.

**Project Context:**  
- **Language/Framework:** Go with `urfave/cli/v3`.
- **Architecture:** Strict functional programming using `github.com/IBM/fp-go/v2`.
- **Constraint:** NO imperative branching (`if`/`else`) in `internal/`. Use `fp-go` combinators (`F.Pipe`, `E.Chain`, `A.Map`, etc.) and `E.Fold` for terminal branching.

**Required Changes:**

1.  **Domain Model (`internal/containerbuild/input.go`):**
    *   Update the `Input` struct to include a `BuildArgs map[string]string` field to store version overrides and generic build arguments.

2.  **CLI Definition (`internal/containerbuild/command.go`):**
    *   Add the following flags to the `build` command:
        *   `--golangci-lint-version` (string, default: "2.11.4")
        *   `--crush-version` (string, default: "latest")
        *   `--gotestsum-version` (string, default: "latest")
        *   `--build-arg` (string slice/repeatable, for generic key=value pairs)
    *   Update `InputFromCommand` to extract these flags and populate the `Input.BuildArgs` map.

3.  **Argument Rendering (`internal/containerbuild/args.go`):**
    *   Modify the function that renders the `container build` command.
    *   For every entry in `Input.BuildArgs`, append two elements to the argv slice: `--build-arg` and `KEY=VALUE`.
    *   Ensure the implementation uses `A.Chain` or `A.Flatten` to maintain the functional style.

4.  **Dockerfile (`Dockerfile`):**
    *   Add `ARG` declarations at the top for `GOLANGCI_LINT_VERSION`, `CRUSH_VERSION`, and `GOTESTSUM_VERSION`.
    *   Update the `RUN` instructions to use these variables instead of hardcoded values.

**Verification Checklist:**
- [ ] No `if` statements or `for` loops in `internal/containerbuild`.
- [ ] Unit tests in `args_test.go` verify that the new flags correctly appear as `--build-arg` in the final command.
- [ ] `golangci-lint run ./...` passes without style or complexity violations.
