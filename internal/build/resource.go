package build

import (
	"errors"
	"os"

	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	IOEF "github.com/IBM/fp-go/v2/ioeither/file"
)

// DockerfileResource pairs a resolved Dockerfile path with its cleanup
// IOEither. For file-based builds cleanup is a nop; for embedded builds
// it removes the temp file via IOEF.Remove.
type DockerfileResource struct {
	Path    string
	Release IOE.IOEither[error, string]
}

var nopRelease = IOE.Of[error]("")

// FileResolver validates that path is non-blank, then wraps it in a
// DockerfileResource with nop cleanup.
func FileResolver(path string) IOE.IOEither[error, DockerfileResource] {
	return F.Pipe1(
		F.Pipe2(
			path,
			E.FromPredicate(isNonBlank, func(string) error {
				return errors.New("dockerfile path is required")
			}),
			E.Map[error](func(p string) DockerfileResource {
				return DockerfileResource{Path: p, Release: nopRelease}
			}),
		),
		IOE.FromEither[error, DockerfileResource],
	)
}

// writeContentAndReturnName writes the embedded content to the file and
// returns the file's Name(). This is the Kleisli arrow passed to
// IOEF.Write — the bracket guarantees the file is closed even if the
// write fails.
func writeContentAndReturnName(
	content []byte,
) func(*os.File) IOE.IOEither[error, string] {
	return func(f *os.File) IOE.IOEither[error, string] {
		return F.Pipe1(
			IOE.TryCatchError(func() (int, error) {
				return f.Write(content)
			}),
			IOE.Map[error](func(int) string {
				return f.Name()
			}),
		)
	}
}

// toDockerfileResource constructs a DockerfileResource from the temp file name.
func toDockerfileResource(name string) DockerfileResource {
	return DockerfileResource{Path: name, Release: IOEF.Remove(name)}
}

// EmbeddedResolver writes the compile-time embedded Dockerfile content to a
// temp file and returns a DockerfileResource whose Release removes that file.
func EmbeddedResolver() IOE.IOEither[error, DockerfileResource] {
	content := []byte(embeddedDockerfile)
	acquire := IOEF.CreateTemp("", "Dockerfile-*")
	writeFile := IOEF.Write[string](acquire)(writeContentAndReturnName(content))

	return F.Pipe1(
		writeFile,
		IOE.Map[error](toDockerfileResource),
	)
}
