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
		APIKey:  "test-key",
	}

	spec := RenderCommand(resolved)

	require.Equal("docker", spec.Bin)
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
		APIKey:  "test-key",
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
		Workdir: WorkspaceTarget,
		APIKey:  "test-key",
	}

	spec := RenderCommand(resolved)

	require.Contains(spec.Args, "--workdir")
	require.Contains(spec.Args, WorkspaceTarget)
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
		APIKey:  "test-key",
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
		APIKey:  "test-key",
	}

	spec := RenderCommand(resolved)

	require.Contains(spec.Args, "--env")
	require.Contains(spec.Args, "CRUSH_GLOBAL_CONFIG="+ConfigTarget)
	require.Contains(spec.Args, "CRUSH_GLOBAL_DATA="+DataTarget)
	require.Contains(spec.Args, "OPENROUTER_API_KEY=test-key")
}

func TestRenderCommand_GitHubTokenPresent(t *testing.T) {
	require := require.New(t)
	resolved := ResolvedInput{
		ImageName:     "crusher:latest",
		ContainerName: "test-container",
		Mounts: []MountSpec{
			{HostPath: "/host/config", TargetPath: ConfigTarget, Readonly: true},
			{HostPath: "/host/data", TargetPath: DataTarget, Readonly: false},
		},
		Workdir:     "",
		APIKey:      "test-key",
		GitHubToken: "ghp_abc123",
	}

	spec := RenderCommand(resolved)

	require.Contains(spec.Args, "GITHUB_TOKEN=ghp_abc123")
}

func TestRenderCommand_GitHubTokenOmitted(t *testing.T) {
	require := require.New(t)
	resolved := ResolvedInput{
		ImageName:     "crusher:latest",
		ContainerName: "test-container",
		Mounts: []MountSpec{
			{HostPath: "/host/config", TargetPath: ConfigTarget, Readonly: true},
			{HostPath: "/host/data", TargetPath: DataTarget, Readonly: false},
		},
		Workdir:     "",
		APIKey:      "test-key",
		GitHubToken: "",
	}

	spec := RenderCommand(resolved)

	for _, arg := range spec.Args {
		require.NotContains(arg, "GITHUB_TOKEN")
	}
}

func TestRenderCommand_AlwaysHasInteractiveAndTTY(t *testing.T) {
	require := require.New(t)
	resolved := ResolvedInput{
		ImageName:     "crusher:latest",
		ContainerName: "test-container",
		Mounts: []MountSpec{
			{HostPath: "/host/config", TargetPath: ConfigTarget, Readonly: true},
			{HostPath: "/host/data", TargetPath: DataTarget, Readonly: false},
		},
		Workdir: "",
		APIKey:  "test-key",
	}

	spec := RenderCommand(resolved)

	require.Contains(spec.Args, "--interactive")
	require.Contains(spec.Args, "--tty")
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
		APIKey:  "test-key",
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
