package build

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	IOE "github.com/IBM/fp-go/v2/ioeither"
)

func Execute(ctx context.Context, spec CommandSpec) IOE.IOEither[error, struct{}] {
	return IOE.TryCatchError(func() (struct{}, error) {
		cmd := exec.CommandContext(ctx, spec.Name, spec.Args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return struct{}{}, fmt.Errorf("container build failed: %w", err)
		}
		return struct{}{}, nil
	})
}
