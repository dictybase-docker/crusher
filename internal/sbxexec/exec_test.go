package sbxexec

import (
	"context"
	"errors"
	"os/exec"
	"testing"
	"time"

	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	FP "github.com/dictybase-docker/crusher/internal/fp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRunSbxCommand_CancelledContextKillsProcess proves that cancelling
// spec.Ctx stops the spawned subprocess — the regression called out in
// docs/opencode-sbx-review.md §1. We run `sleep 30`, cancel almost
// immediately, and assert the command returns a Left within a short
// deadline (well below 30s).
func TestRunSbxCommand_CancelledContextKillsProcess(t *testing.T) {
	require := require.New(t)

	// Confirm sleep is available; a missing binary would otherwise surface as
	// an unrelated LookPath failure and mask the cancellation behaviour.
	if _, err := exec.LookPath("sleep"); err != nil {
		t.Skipf("sleep binary not present: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	spec := CommandSpec{
		Ctx:  ctx,
		Bin:  "sleep",
		Args: []string{"30"},
	}

	type eval struct {
		either E.Either[error, F.Void]
	}

	done := make(chan eval, 1)
	go func() {
		done <- eval{either: FP.ToEither[error, F.Void](RunSbxCommand(spec))}
	}()

	cancel()

	select {
	case res := <-done:
		assert.True(t, E.IsLeft(res.either), "expected Left error from cancelled context")

		err := E.ToError(res.either)
		require.Error(err)

		var ctxErr *exec.ExitError
		if errors.As(err, &ctxErr) {
			t.Logf("exit error captured: %v", ctxErr)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("RunSbxCommand did not return after context cancellation")
	}
}

// TestRunSbxCommand_MissingBinaryReturnsLeft confirms a missing binary
// surfaces as a Left (LookPath failure path), not a panic.
func TestRunSbxCommand_MissingBinaryReturnsLeft(t *testing.T) {
	spec := CommandSpec{
		Ctx:  context.Background(),
		Bin:  "this-binary-does-not-exist-xyz",
		Args: []string{},
	}

	// Unused locally but kept to document the type parameter convention of
	// the codebase (FP.ToEither is generic).
	_ = IOE.Of[error](F.VOID)

	either := FP.ToEither[error, F.Void](RunSbxCommand(spec))
	assert.True(t, E.IsLeft(either))
}
