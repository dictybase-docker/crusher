package build

import (
	"fmt"
	"os"
	"os/exec"

	IOE "github.com/IBM/fp-go/v2/ioeither"
)

func Execute(r Input) IOE.IOEither[error, struct{}] {
	return IOE.TryCatchError(func() (struct{}, error) {
		cmd := exec.CommandContext(r.Ctx, containerBinary)
		cmd.Args = append(cmd.Args, r.Args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return struct{}{}, fmt.Errorf("container build failed: %w", err)
		}
		return struct{}{}, nil
	})
}
