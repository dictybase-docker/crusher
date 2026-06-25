package sbxexec

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	F "github.com/IBM/fp-go/v2/function"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	O "github.com/IBM/fp-go/v2/option"
	Str "github.com/IBM/fp-go/v2/string"
)

// stdinReader converts an optional stdin string into a reader, returning nil
// for the empty case so exec.Cmd leaves stdin untouched. Kept point-free via
// Option combinators to honour the no-imperative-branching rule.
func stdinReader(s string) io.Reader {
	return F.Pipe2(
		s,
		O.FromPredicate(Str.IsNonEmpty),
		O.Fold(
			F.Constant[io.Reader](nil),
			func(v string) io.Reader { return strings.NewReader(v) },
		),
	)
}

// RunSbxCommand looks up the sbx binary then runs the command under
// spec.Ctx, optionally piping Stdin. The context governs cancellation and
// deadlines: when it is cancelled, the spawned process is killed.
// Returns Left on lookup failure or non-zero exit.
//
// exec.CommandContext is used instead of a bare &exec.Cmd{} so that
// spec.Ctx actually drives process lifetime. gosec rule G204 flags any
// subprocess launched with non-constant arguments; here spec.Args is built
// from validated CLI input inside this same process (never a raw shell
// string), so the directive is a false positive and is suppressed locally.
func RunSbxCommand(spec CommandSpec) IOE.IOEither[error, F.Void] {
	return F.Pipe2(
		IOE.TryCatchError(func() (string, error) {
			return exec.LookPath(spec.Bin)
		}),
		IOE.Chain(func(bin string) IOE.IOEither[error, F.Void] {
			return IOE.TryCatchError(func() (F.Void, error) {
				//nolint:gosec // G204: args are validated, built in-process, no shell
				cmd := exec.CommandContext(spec.Ctx, bin, spec.Args...)
				cmd.Stdin = stdinReader(spec.Stdin)
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr

				return F.VOID, cmd.Run()
			})
		}),
		IOE.MapLeft[F.Void](func(err error) error {
			return fmt.Errorf("sbx command failed: %w", err)
		}),
	)
}
