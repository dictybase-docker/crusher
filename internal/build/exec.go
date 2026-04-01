package build

import (
	"context"
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
func useResource(r Input) func(DockerfileResource) IOE.IOEither[error, struct{}] {
	return func(res DockerfileResource) IOE.IOEither[error, struct{}] {
		return runProcess(r.Ctx, RenderCommand(r, res.Path))
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
func Execute(r Input) IOE.IOEither[error, struct{}] {
	return F.Pipe1(
		useResource(r),
		IOE.WithResource[struct{}](r.DockerfileSource, releaseResource),
	)
}

// runProcess executes the container binary with the given CommandSpec.
// Split from Execute to satisfy funlen limits.
func runProcess(ctx context.Context, spec CommandSpec) IOE.IOEither[error, struct{}] {
	return IOE.TryCatchError(func() (struct{}, error) {
		cmd := exec.CommandContext(ctx, spec.Bin) //nolint:gosec
		cmd.Args = append(cmd.Args, spec.Args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return struct{}{}, fmt.Errorf("container build failed: %w", err)
		}
		return struct{}{}, nil
	})
}
