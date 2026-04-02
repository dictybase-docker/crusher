package containerbuild

import (
	"fmt"
	"os"
	"os/exec"

	F "github.com/IBM/fp-go/v2/function"
	IOE "github.com/IBM/fp-go/v2/ioeither"
)

// releaseResource is the Kleisli arrow passed to ioeither.WithResource as the
// release callback. It delegates to the resource's Release field (nop for
// file-based, IOEF.Remove for embedded).
func releaseResource(res DockerfileResource) IOE.IOEither[error, string] {
	return res.Release
}

// useResource builds the use Kleisli arrow for ioeither.WithResource.
// It renders the argv from the resolved path and runs the container process.
func useResource(r Input) func(DockerfileResource) IOE.IOEither[error, F.Void] {
	return func(res DockerfileResource) IOE.IOEither[error, F.Void] {
		return runProcess(RenderCommand(r, res.Path))
	}
}

// Execute orchestrates the full build lifecycle using ioeither.WithResource:
//
//  1. acquire — run DockerfileSource to get a DockerfileResource
//  2. use    — render the argv and run the container process
//  3. release — run res.Release (nop for file, IOEF.Remove for embedded)
//
// release is guaranteed to run even when the container process fails.
//
// ioeither.WithResource returns a Kleisli[E, Kleisli[E, R, A], A], so we
// call it with the use Kleisli to get the final IOEither.
func Execute(r Input) IOE.IOEither[error, F.Void] {
	return F.Pipe1(
		useResource(r),
		IOE.WithResource[F.Void](r.DockerfileSource, releaseResource),
	)
}

// runProcess executes the container binary with the given CommandSpec.
// Split from Execute to satisfy funlen limits.
func runProcess(spec CommandSpec) IOE.IOEither[error, F.Void] {
	return F.Pipe1(
		IOE.TryCatchError(func() (F.Void, error) {
			cmd := &exec.Cmd{
				Path:   spec.Bin,
				Args:   append([]string{spec.Bin}, spec.Args...),
				Stdout: os.Stdout,
				Stderr: os.Stderr,
			}
			return F.VOID, cmd.Run()
		}),
		IOE.MapLeft[F.Void](func(err error) error {
			return fmt.Errorf("container build failed: %w", err)
		}),
	)
}
