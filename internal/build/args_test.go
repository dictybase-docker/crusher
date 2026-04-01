package build

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRenderCommand_DefaultInput(t *testing.T) {
	require := require.New(t)
	req := Input{
		Name: "myapp",
		Tags: []string{"latest"},
	}

	spec := RenderCommand(req, "Dockerfile")

	require.Equal(containerBinary, spec.Bin)

	expected := containerBinary + " build --file Dockerfile --tag myapp:latest ."
	actual := spec.Bin + " " + strings.Join(spec.Args, " ")
	require.Equal(expected, actual)
}

func TestRenderCommand_RepeatedTags(t *testing.T) {
	require := require.New(t)
	req := Input{
		Name: "myapp",
		Tags: []string{"latest", "stable", "v1.0.0"},
	}

	spec := RenderCommand(req, "Dockerfile")

	require.Equal(containerBinary, spec.Bin)

	expected := containerBinary + " build --file Dockerfile --tag myapp:latest --tag myapp:stable --tag myapp:v1.0.0 ."
	actual := spec.Bin + " " + strings.Join(spec.Args, " ")
	require.Equal(expected, actual)
}

func TestRenderCommand_DockerfileOverride(t *testing.T) {
	require := require.New(t)
	req := Input{
		Name: "myapp",
		Tags: []string{"latest"},
	}

	spec := RenderCommand(req, "docker/Prod.Dockerfile")

	require.Equal(containerBinary, spec.Bin)

	expected := containerBinary + " build --file docker/Prod.Dockerfile --tag myapp:latest ."
	actual := spec.Bin + " " + strings.Join(spec.Args, " ")
	require.Equal(expected, actual)
}

func TestRenderCommand_FinalArgIsBuildContext(t *testing.T) {
	require := require.New(t)
	req := Input{
		Name: "myapp",
		Tags: []string{"latest"},
	}

	spec := RenderCommand(req, "Dockerfile")

	require.NotEmpty(spec.Args)

	lastArg := spec.Args[len(spec.Args)-1]
	require.Equal(".", lastArg)
}

func TestRenderCommand_ArgsOrder(t *testing.T) {
	require := require.New(t)
	req := Input{
		Name: "myapp",
		Tags: []string{"latest", "stable"},
	}

	spec := RenderCommand(req, "Dockerfile")

	require.GreaterOrEqual(len(spec.Args), 7)

	require.Equal("build", spec.Args[0])
	require.Equal("--file", spec.Args[1])
	require.Equal("Dockerfile", spec.Args[2])
	require.Equal("--tag", spec.Args[3])
	require.Equal("myapp:latest", spec.Args[4])
	require.Equal("--tag", spec.Args[5])
	require.Equal("myapp:stable", spec.Args[6])
}
