package containercreate

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRenderCommand_MinimalInput(t *testing.T) {
	require := require.New(t)
	resolved := ResolvedInput{
		ImageName:     "crusher:latest",
		ContainerName: "test-container",
		Mounts: []MountSpec{
			{HostPath: "/host/config", TargetPath: ConfigTarget, Readonly: true},
			{HostPath: "/host/data", TargetPath: DataTarget, Readonly: false},
		},
		Workdir: "",
	}

	spec := RenderCommand(resolved)

	require.Equal("container", spec.Bin)
	require.Equal("create", spec.Args[0])
	require.Contains(spec.Args, "--name")
	require.Contains(spec.Args, "test-container")
	require.Contains(spec.Args, "crusher:latest")
}

func TestRenderCommand_WithWorkdir(t *testing.T) {
	require := require.New(t)
	resolved := ResolvedInput{
		ImageName:     "crusher:latest",
		ContainerName: "test-container",
		Mounts: []MountSpec{
			{HostPath: "/host/config", TargetPath: ConfigTarget, Readonly: true},
			{HostPath: "/host/data", TargetPath: DataTarget, Readonly: false},
			{HostPath: "/host/workspace", TargetPath: WorkspaceTarget, Readonly: false},
		},
		Workdir: WorkspaceTarget,
	}

	spec := RenderCommand(resolved)

	require.Contains(spec.Args, "--workdir")
	require.Contains(spec.Args, WorkspaceTarget)
}

func TestRenderCommand_WithoutWorkdir(t *testing.T) {
	require := require.New(t)
	resolved := ResolvedInput{
		ImageName:     "crusher:latest",
		ContainerName: "test-container",
		Mounts: []MountSpec{
			{HostPath: "/host/config", TargetPath: ConfigTarget, Readonly: true},
			{HostPath: "/host/data", TargetPath: DataTarget, Readonly: false},
		},
		Workdir: "",
	}

	spec := RenderCommand(resolved)

	require.NotContains(spec.Args, "--workdir")
}

func TestRenderCommand_ImageNameIsLastArg(t *testing.T) {
	require := require.New(t)
	resolved := ResolvedInput{
		ImageName:     "myimage:v1",
		ContainerName: "test-container",
		Mounts: []MountSpec{
			{HostPath: "/host/config", TargetPath: ConfigTarget, Readonly: true},
			{HostPath: "/host/data", TargetPath: DataTarget, Readonly: false},
		},
		Workdir: "",
	}

	spec := RenderCommand(resolved)

	require.NotEmpty(spec.Args)
	require.Equal("myimage:v1", spec.Args[len(spec.Args)-1])
}

func TestRenderCommand_EnvVarsPresent(t *testing.T) {
	require := require.New(t)
	resolved := ResolvedInput{
		ImageName:     "crusher:latest",
		ContainerName: "test-container",
		Mounts: []MountSpec{
			{HostPath: "/host/config", TargetPath: ConfigTarget, Readonly: true},
			{HostPath: "/host/data", TargetPath: DataTarget, Readonly: false},
		},
		Workdir: "",
	}

	spec := RenderCommand(resolved)

	require.Contains(spec.Args, "--env")
	require.Contains(spec.Args, "CRUSH_GLOBAL_CONFIG="+ConfigTarget)
	require.Contains(spec.Args, "CRUSH_GLOBAL_DATA="+DataTarget)
}

func TestBuildArgs_ArgsOrder(t *testing.T) {
	require := require.New(t)
	resolved := ResolvedInput{
		ImageName:     "crusher:latest",
		ContainerName: "test-container",
		Mounts: []MountSpec{
			{HostPath: "/host/config", TargetPath: ConfigTarget, Readonly: true},
			{HostPath: "/host/data", TargetPath: DataTarget, Readonly: false},
		},
		Workdir: "",
	}

	spec := RenderCommand(resolved)

	require.GreaterOrEqual(len(spec.Args), 3)
	require.Equal("create", spec.Args[0])

	nameIdx := -1
	for i, arg := range spec.Args {
		if arg == "--name" {
			nameIdx = i
			break
		}
	}
	require.GreaterOrEqual(nameIdx, 0)
	require.Equal("test-container", spec.Args[nameIdx+1])
}
