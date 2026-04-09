package containercreate

import (
	"fmt"
	"os"
	"os/exec"

	F "github.com/IBM/fp-go/v2/function"
	IOE "github.com/IBM/fp-go/v2/ioeither"
)

// Execute runs the container create command and returns the result.
func Execute(r ResolvedInput) IOE.IOEither[error, ContainerResult] {
	return F.Pipe3(
		r,
		RenderCommand,
		runProcess,
		IOE.Map[error](func(F.Void) ContainerResult {
			return ContainerResult{Name: r.ContainerName}
		}),
	)
}

// StartContainer runs the container start command for a created container.
func StartContainer(result ContainerResult) IOE.IOEither[error, ContainerResult] {
	return F.Pipe2(
		CommandSpec{
			Bin:  containerBinary,
			Args: []string{"start", result.Name},
		},
		runProcess,
		IOE.Map[error](func(F.Void) ContainerResult {
			return result
		}),
	)
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
			return fmt.Errorf("container command failed: %w", err)
		}),
	)
}
