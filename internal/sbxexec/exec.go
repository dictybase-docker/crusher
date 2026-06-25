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
		O.Fold[string, io.Reader](
			F.Constant[io.Reader](nil),
			func(v string) io.Reader { return strings.NewReader(v) },
		),
	)
}

// RunSbxCommand looks up the sbx binary then runs the command, optionally
// piping Stdin. Returns Left on lookup failure or non-zero exit.
func RunSbxCommand(spec CommandSpec) IOE.IOEither[error, F.Void] {
	return F.Pipe2(
		IOE.TryCatchError(func() (string, error) {
			return exec.LookPath(spec.Bin)
		}),
		IOE.Chain(func(bin string) IOE.IOEither[error, F.Void] {
			return IOE.TryCatchError(func() (F.Void, error) {
				cmd := &exec.Cmd{
					Path:   bin,
					Args:   append([]string{bin}, spec.Args...),
					Stdin:  stdinReader(spec.Stdin),
					Stdout: os.Stdout,
					Stderr: os.Stderr,
				}

				return F.VOID, cmd.Run()
			})
		}),
		IOE.MapLeft[F.Void](func(err error) error {
			return fmt.Errorf("sbx command failed: %w", err)
		}),
	)
}
