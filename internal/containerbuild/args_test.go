package containerbuild

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	testName      = "myapp"
	defaultTag    = "latest"
	stableTag     = "stable"
	glvKey        = "GOLANGCI_LINT_VERSION"
	glvDefVal     = "2.11.4"
	testImageName = "testimage"
	v100Tag       = "v1.0.0"
)

func TestRenderCommand_DefaultInput(t *testing.T) {
	require := require.New(t)
	req := Input{
		Name:      testName,
		Tags:      []string{defaultTag},
		BuildArgs: map[string]string{},
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
		Name:      testName,
		Tags:      []string{defaultTag, stableTag, v100Tag},
		BuildArgs: map[string]string{},
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
		Name:      testName,
		Tags:      []string{defaultTag},
		BuildArgs: map[string]string{},
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
		Name:      testName,
		Tags:      []string{defaultTag},
		BuildArgs: map[string]string{},
	}

	spec := RenderCommand(req, "Dockerfile")

	require.NotEmpty(spec.Args)

	lastArg := spec.Args[len(spec.Args)-1]
	require.Equal(".", lastArg)
}

func TestRenderCommand_ArgsOrder(t *testing.T) {
	require := require.New(t)
	req := Input{
		Name:      testName,
		Tags:      []string{defaultTag, stableTag},
		BuildArgs: map[string]string{},
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

func TestRenderCommand_BuildArgs(t *testing.T) {
	require := require.New(t)
	req := Input{
		Name: testName,
		Tags: []string{defaultTag},
		BuildArgs: map[string]string{
			glvKey:           "2.12.0",
			"CRUSH_VERSION":  v100Tag,
			"MOXIDE_VERSION": "v0.25.10",
		},
	}

	spec := RenderCommand(req, "Dockerfile")

	require.Contains(spec.Args, "--build-arg")
	require.Contains(spec.Args, "GOLANGCI_LINT_VERSION=2.12.0")
	require.Contains(spec.Args, "CRUSH_VERSION=v1.0.0")
	require.Contains(spec.Args, "MOXIDE_VERSION=v0.25.10")
}

func TestRenderCommand_BuildArgs_SortedOrder(t *testing.T) {
	require := require.New(t)
	req := Input{
		Name: testName,
		Tags: []string{defaultTag},
		BuildArgs: map[string]string{
			glvKey:              glvDefVal,
			"CRUSH_VERSION":     defaultTag,
			"GOTESTSUM_VERSION": defaultTag,
			"MOXIDE_VERSION":    defaultTag,
			"RTK_VERSION":       defaultTag,
			"SEM_VERSION":       defaultTag,
		},
	}

	spec := RenderCommand(req, "Dockerfile")

	var buildArgValues []string

	for i, arg := range spec.Args {
		if arg == "--build-arg" && i+1 < len(spec.Args) {
			buildArgValues = append(buildArgValues, spec.Args[i+1])
		}
	}

	require.Len(buildArgValues, 6)
	require.Equal("CRUSH_VERSION=latest", buildArgValues[0])
	require.Equal("GOLANGCI_LINT_VERSION=2.11.4", buildArgValues[1])
	require.Equal("GOTESTSUM_VERSION=latest", buildArgValues[2])
	require.Equal("MOXIDE_VERSION=latest", buildArgValues[3])
	require.Equal("RTK_VERSION=latest", buildArgValues[4])
	require.Equal("SEM_VERSION=latest", buildArgValues[5])
}

func TestRenderCommand_BuildArgs_Position(t *testing.T) {
	require := require.New(t)
	req := Input{
		Name: testName,
		Tags: []string{defaultTag},
		BuildArgs: map[string]string{
			glvKey: glvDefVal,
		},
	}

	spec := RenderCommand(req, "Dockerfile")

	require.Equal(".", spec.Args[len(spec.Args)-1])

	buildArgIdx := -1

	for i, arg := range spec.Args {
		if arg == "GOLANGCI_LINT_VERSION=2.11.4" {
			buildArgIdx = i
			break
		}
	}

	require.Less(buildArgIdx, len(spec.Args)-1)
}

func TestRenderCommand_EmptyBuildArgs(t *testing.T) {
	require := require.New(t)
	req := Input{
		Name:      testName,
		Tags:      []string{defaultTag},
		BuildArgs: map[string]string{},
	}

	spec := RenderCommand(req, "Dockerfile")

	require.NotContains(spec.Args, "--build-arg")
}
