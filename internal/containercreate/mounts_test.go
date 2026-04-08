package containercreate

import (
	"os"
	"testing"

	A "github.com/IBM/fp-go/v2/array"
	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	"github.com/stretchr/testify/require"
)

func TestRenderMount_ReadonlyMount(t *testing.T) {
	require := require.New(t)
	mount := MountSpec{
		HostPath:   "/host/path",
		TargetPath: "/container/path",
		Readonly:   true,
	}

	result := renderMount(mount)

	require.Contains(result, "type=bind")
	require.Contains(result, "source=/host/path")
	require.Contains(result, "target=/container/path")
	require.Contains(result, ",readonly")
}

func TestRenderMount_ReadwriteMount(t *testing.T) {
	require := require.New(t)
	mount := MountSpec{
		HostPath:   "/host/path",
		TargetPath: "/container/path",
		Readonly:   false,
	}

	result := renderMount(mount)

	require.Contains(result, "type=bind")
	require.Contains(result, "source=/host/path")
	require.Contains(result, "target=/container/path")
	require.NotContains(result, "readonly")
}

func TestRenderMountArgs_ReturnsPair(t *testing.T) {
	require := require.New(t)
	mount := MountSpec{
		HostPath:   "/host/path",
		TargetPath: "/container/path",
		Readonly:   true,
	}

	result := renderMountArgs(mount)

	require.Len(result, 2)
	require.Equal("--mount", result[0])
	require.Contains(result[1], "type=bind")
}

func TestRenderAllMounts_MultipleMounts(t *testing.T) {
	require := require.New(t)
	mounts := []MountSpec{
		{HostPath: "/host/config", TargetPath: ConfigTarget, Readonly: true},
		{HostPath: "/host/data", TargetPath: DataTarget, Readonly: false},
	}

	result := F.Pipe1(
		mounts,
		A.Chain(renderMountArgs),
	)

	require.Len(result, 4)
	require.Equal("--mount", result[0])
	require.Equal("--mount", result[2])
}

func TestRenderAllMounts_EmptyMounts(t *testing.T) {
	require := require.New(t)
	mounts := []MountSpec{}

	result := F.Pipe1(
		mounts,
		A.Chain(renderMountArgs),
	)

	require.Empty(result)
}

func TestRenderEnvVars_ContainsCrushPaths(t *testing.T) {
	require := require.New(t)

	result := renderEnvVars()

	require.Len(result, 4)
	require.Equal("--env", result[0])
	require.Contains(result[1], "CRUSH_GLOBAL_CONFIG="+ConfigTarget)
	require.Equal("--env", result[2])
	require.Contains(result[3], "CRUSH_GLOBAL_DATA="+DataTarget)
}

func TestRenderMount_SpecialCharacters(t *testing.T) {
	require := require.New(t)
	mount := MountSpec{
		HostPath:   "/host/path with spaces",
		TargetPath: "/container/path-with-dashes",
		Readonly:   true,
	}

	result := renderMount(mount)

	require.Contains(result, "source=/host/path with spaces")
	require.Contains(result, "target=/container/path-with-dashes")
}

func TestConfigMount_IsReadonly(t *testing.T) {
	require := require.New(t)
	configDir := t.TempDir()
	dataDir := t.TempDir()

	input := Input{
		ConfigPath: configDir,
		DataPath:   dataDir,
	}

	result := F.Pipe2(
		input,
		NormalizeInput,
		ValidateInput,
	)
	require.True(E.IsRight(result))

	resolved := F.Pipe1(
		result,
		E.Fold(func(error) ResolvedInput { return ResolvedInput{} }, F.Identity[ResolvedInput]),
	)

	var configMount *MountSpec
	for i := range resolved.Mounts {
		if resolved.Mounts[i].TargetPath == ConfigTarget {
			configMount = &resolved.Mounts[i]
			break
		}
	}

	require.NotNil(configMount, "config mount should exist")
	require.True(configMount.Readonly, "config mount should be read-only")
}

func TestDataMount_IsReadwrite(t *testing.T) {
	require := require.New(t)
	configDir := t.TempDir()
	dataDir := t.TempDir()

	input := Input{
		ConfigPath: configDir,
		DataPath:   dataDir,
	}

	result := F.Pipe2(
		input,
		NormalizeInput,
		ValidateInput,
	)
	require.True(E.IsRight(result))

	resolved := F.Pipe1(
		result,
		E.Fold(func(error) ResolvedInput { return ResolvedInput{} }, F.Identity[ResolvedInput]),
	)

	var dataMount *MountSpec
	for i := range resolved.Mounts {
		if resolved.Mounts[i].TargetPath == DataTarget {
			dataMount = &resolved.Mounts[i]
			break
		}
	}

	require.NotNil(dataMount, "data mount should exist")
	require.False(dataMount.Readonly, "data mount should be read-write")
}

func TestAdditionalVolume_IsReadonly(t *testing.T) {
	require := require.New(t)
	configDir := t.TempDir()
	dataDir := t.TempDir()
	parent := t.TempDir()
	volDir := parent + "/myproject"
	require.NoError(os.MkdirAll(volDir, 0o755))

	input := Input{
		ConfigPath: configDir,
		DataPath:   dataDir,
		Volumes:    []string{volDir},
	}

	result := F.Pipe2(
		input,
		NormalizeInput,
		ValidateInput,
	)
	require.True(E.IsRight(result))

	resolved := F.Pipe1(
		result,
		E.Fold(func(error) ResolvedInput { return ResolvedInput{} }, F.Identity[ResolvedInput]),
	)

	expectedTarget := ContainerHome + "/myproject"
	var volumeMount *MountSpec
	for i := range resolved.Mounts {
		if resolved.Mounts[i].TargetPath == expectedTarget {
			volumeMount = &resolved.Mounts[i]
			break
		}
	}

	require.NotNil(volumeMount, "volume mount should exist")
	require.True(volumeMount.Readonly, "additional volume mount should be read-only")
}
