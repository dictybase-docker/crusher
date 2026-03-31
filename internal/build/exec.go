package build

import (
	"fmt"
	"os"
	"os/exec"

	IOE "github.com/IBM/fp-go/v2/ioeither"
)

func Execute(r Request) IOE.IOEither[error, struct{}] {
	return IOE.TryCatchError(func() (struct{}, error) {
		//nolint:gosec // G204: This CLI intentionally wraps the container binary with user-provided args
		cmd := exec.CommandContext(r.ctx, r.Name, r.Args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return struct{}{}, fmt.Errorf("container build failed: %w", err)
		}
		return struct{}{}, nil
	})
}
