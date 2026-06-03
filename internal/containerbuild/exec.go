package containerbuild

import (
	"fmt"
	"os"
	"os/exec"

	F "github.com/IBM/fp-go/v2/function"
	IOE "github.com/IBM/fp-go/v2/ioeither"
)

// processRunner is a type alias for a subprocess runner, enabling injection
// of test doubles.
type processRunner func(spec CommandSpec) IOE.IOEither[error, F.Void]

// Execute orchestrates the full build lifecycle using ioeither.WithResource.
// Delegates to executeWith for testability.
func Execute(r Input) IOE.IOEither[error, F.Void] {
	return executeWith(runProcess, r)
}

// executeWith is the internal parameterized variant that accepts a
// processRunner, enabling unit tests to inject a stub.
func executeWith(run processRunner, r Input) IOE.IOEither[error, F.Void] {
	return F.Pipe1(
		useResource(run, r),
		IOE.WithResource[F.Void](r.DockerfileSource, releaseResource),
	)
}

// releaseResource is the Kleisli arrow passed to ioeither.WithResource as the
// release callback.
var releaseResource = func(res DockerfileResource) IOE.IOEither[error, string] {
	return res.Release
}

// useResource builds the use Kleisli arrow for ioeither.WithResource.
func useResource(run processRunner, r Input) func(DockerfileResource) IOE.IOEither[error, F.Void] {
	return func(res DockerfileResource) IOE.IOEither[error, F.Void] {
		return run(RenderCommand(r, res.Path))
	}
}

// runProcess executes the container binary with the given CommandSpec.
func runProcess(spec CommandSpec) IOE.IOEither[error, F.Void] {
	return F.Pipe2(
		IOE.TryCatchError(func() (string, error) {
			return exec.LookPath(spec.Bin)
		}),
		IOE.Chain(func(bin string) IOE.IOEither[error, F.Void] {
			return IOE.TryCatchError(func() (F.Void, error) {
				cmd := &exec.Cmd{
					Path:   bin,
					Args:   append([]string{bin}, spec.Args...),
					Stdout: os.Stdout,
					Stderr: os.Stderr,
				}
				return F.VOID, cmd.Run()
			})
		}),
		IOE.MapLeft[F.Void](func(err error) error {
			return fmt.Errorf("container build failed: %w", err)
		}),
	)
}
