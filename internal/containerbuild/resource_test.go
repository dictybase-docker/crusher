package containerbuild

import (
	"os"
	"testing"

	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	"github.com/stretchr/testify/require"
)

func TestFileResolver_EmptyPath(t *testing.T) {
	require := require.New(t)

	result := FileResolver("")()

	require.True(E.IsLeft(result), "expected Left for empty path")

	err := F.Pipe1(
		result,
		E.Fold(
			F.Identity[error],
			func(DockerfileResource) error { return nil },
		),
	)
	require.EqualError(err, "dockerfile path is required")
}

func TestFileResolver_BlankPath(t *testing.T) {
	require := require.New(t)

	result := FileResolver("   ")()

	require.True(E.IsLeft(result), "expected Left for blank path")
}

func TestFileResolver_ValidPath(t *testing.T) {
	require := require.New(t)

	result := FileResolver("Dockerfile")()

	require.True(E.IsRight(result), "expected Right for valid path")

	res := F.Pipe1(
		result,
		E.Fold(
			func(error) DockerfileResource { return DockerfileResource{} },
			F.Identity[DockerfileResource],
		),
	)
	require.Equal("Dockerfile", res.Path)
}

func TestFileResolver_ReleaseIsNop(t *testing.T) {
	require := require.New(t)

	result := FileResolver("Dockerfile")()
	require.True(E.IsRight(result))

	res := E.GetOrElse(func(error) DockerfileResource {
		return DockerfileResource{}
	})(result)

	releaseResult := res.Release()
	require.True(E.IsRight(releaseResult), "file resolver release must succeed (nop)")
}

func TestEmbeddedResolver_WritesEmbeddedContent(t *testing.T) {
	require := require.New(t)

	result := EmbeddedResolver()()
	require.True(E.IsRight(result), "expected Right from EmbeddedResolver")

	res := E.GetOrElse(func(error) DockerfileResource {
		return DockerfileResource{}
	})(result)

	defer func() { _ = res.Release() }()

	content, err := os.ReadFile(res.Path)
	require.NoError(err)
	require.Equal(embeddedDockerfile, string(content))
}

func TestEmbeddedResolver_ReleaseCleansUp(t *testing.T) {
	require := require.New(t)

	result := EmbeddedResolver()()
	require.True(E.IsRight(result))

	res := E.GetOrElse(func(error) DockerfileResource {
		return DockerfileResource{}
	})(result)

	path := res.Path
	require.FileExists(path, "temp file must exist before release")

	releaseResult := res.Release()
	require.True(E.IsRight(releaseResult))
	require.NoFileExists(path, "temp file must be removed after release")
}
