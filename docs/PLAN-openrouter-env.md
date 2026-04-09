# Plan: Expose API Key as OPENROUTER_API_KEY Environment Variable

## Problem

The `--api-key` (`-k`) flag is collected in `Input.APIKey` but is **never rendered** into the `container create` command. The API key needs to be passed into the container as the `OPENROUTER_API_KEY` environment variable so Crush can authenticate with OpenRouter.

## Current State

- `Input.APIKey` exists and is populated from `cmd.String("api-key")` in `InputFromCommand()`
- `renderEnvVars()` currently returns only two env vars: `CRUSH_GLOBAL_CONFIG` and `CRUSH_GLOBAL_DATA`
- `ResolvedInput` struct does **not** include `APIKey` — it's dropped in `buildResolvedInput()`
- The API key never reaches `RenderCommand()` or the container process

## Target State

The API key value from `--api-key` is passed through the validation pipeline and rendered as `--env OPENROUTER_API_KEY=<value>` in the `container create` command.

---

## Changes

### 1. `internal/containercreate/input.go` — Add `APIKey` to `ResolvedInput`

```diff
 type ResolvedInput struct {
 	ImageName     string
 	ContainerName string
 	Mounts        []MountSpec
 	Workdir       string
+	APIKey        string
 }
```

**Why**: `ResolvedInput` is the validated type that `RenderCommand` consumes. The API key must be present here to be rendered into the command.

---

### 2. `internal/containercreate/validate.go` — Thread `APIKey` into `buildResolvedInput`

```diff
 func buildResolvedInput(input Input) ResolvedInput {
 	...existing pipeline...
 		func(mspec []MountSpec) ResolvedInput {
 			return ResolvedInput{
 				ImageName:     input.ImageName,
 				ContainerName: input.ContainerName,
 				Mounts:        mspec,
 				Workdir:       WorkspaceTarget,
+				APIKey:        input.APIKey,
 			}
 		},
 	)
```

**Why**: `buildResolvedInput` constructs the final `ResolvedInput`. Currently it drops `APIKey`; threading it through preserves the value for rendering.

---

### 3. `internal/containercreate/mounts.go` — Accept `apiKey` parameter, add `OPENROUTER_API_KEY`

```diff
-func renderEnvVars() []string {
+func renderEnvVars(apiKey string) []string {
 	return []string{
 		"--env", fmt.Sprintf("CRUSH_GLOBAL_CONFIG=%s", ConfigTarget),
 		"--env", fmt.Sprintf("CRUSH_GLOBAL_DATA=%s", DataTarget),
+		"--env", fmt.Sprintf("OPENROUTER_API_KEY=%s", apiKey),
 	}
 }
```

**Why**: The function now produces three environment variables. The API key is injected as `OPENROUTER_API_KEY` which Crush reads at runtime.

---

### 4. `internal/containercreate/args.go` — Replace inline env-var concatenations with `renderEnvVars(rinput.APIKey)`

```diff
 func RenderCommand(rinput ResolvedInput) CommandSpec {
-	return F.Pipe7(
+	return F.Pipe6(
 		A.Of("create"),
 		A.Concat([]string{
 			"--name",
 			rinput.ContainerName,
 		}),
 		A.Concat(F.Pipe1(
 			rinput.Mounts,
 			A.Chain(renderMount),
 		)),
-		A.Concat([]string{
-			"--env",
-			fmt.Sprintf("CRUSH_GLOBAL_CONFIG=%s", ConfigTarget),
-		}),
-		A.Concat([]string{
-			"--env",
-			fmt.Sprintf("CRUSH_GLOBAL_DATA=%s", DataTarget),
-		}),
+		A.Concat(renderEnvVars(rinput.APIKey)),
 		A.Concat([]string{"--workdir", rinput.Workdir}),
 		A.Push(rinput.ImageName),
 		func(args []string) CommandSpec {
 			return CommandSpec{Bin: containerBinary, Args: args}
 		},
 	)
 }
```

**Why**: Consolidates two `A.Concat` calls into one, using `renderEnvVars()` as the single source of truth for all env vars. Reduces `F.Pipe7` to `F.Pipe6`. The `fmt` import can be removed since `Sprintf` is no longer used directly in this file.

---

### 5. `internal/containercreate/mounts_test.go` — Update `renderEnvVars` test

```diff
-func TestRenderEnvVars_ContainsCrushPaths(t *testing.T) {
+func TestRenderEnvVars_ContainsAllEnvVars(t *testing.T) {
 	require := require.New(t)
 
-	result := renderEnvVars()
+	result := renderEnvVars("test-api-key-123")
 
-	require.Len(result, 4)
+	require.Len(result, 6)
 	require.Equal("--env", result[0])
 	require.Contains(result[1], "CRUSH_GLOBAL_CONFIG="+ConfigTarget)
 	require.Equal("--env", result[2])
 	require.Contains(result[3], "CRUSH_GLOBAL_DATA="+DataTarget)
+	require.Equal("--env", result[4])
+	require.Contains(result[5], "OPENROUTER_API_KEY=test-api-key-123")
 }
```

---

### 6. `internal/containercreate/args_test.go` — Add `APIKey` to all fixtures, assert `OPENROUTER_API_KEY`

Every `ResolvedInput` test fixture needs an `APIKey` field. The `TestRenderCommand_EnvVarsPresent` test should also verify `OPENROUTER_API_KEY` is present.

---

## Files Changed Summary

| File | Change |
|------|--------|
| `internal/containercreate/input.go` | Add `APIKey string` to `ResolvedInput` |
| `internal/containercreate/validate.go` | Add `APIKey: input.APIKey` in `buildResolvedInput` |
| `internal/containercreate/mounts.go` | `renderEnvVars(apiKey string)` — add `OPENROUTER_API_KEY` |
| `internal/containercreate/args.go` | Replace two env-var `A.Concat` calls with `A.Concat(renderEnvVars(rinput.APIKey))`, reduce pipe arity |
| `internal/containercreate/mounts_test.go` | Update `renderEnvVars` test with API key param, assert `OPENROUTER_API_KEY` |
| `internal/containercreate/args_test.go` | Add `APIKey` to all `ResolvedInput` fixtures, assert `OPENROUTER_API_KEY` present |