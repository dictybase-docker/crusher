package containercreate

import (
	"testing"

	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	"github.com/stretchr/testify/require"
)

func TestValidateInput_MissingConfigPath(t *testing.T) {
	require := require.New(t)
	input := Input{
		ConfigPath: "",
		DataPath:   "/tmp/data",
	}

	result := ValidateInput(input)

	require.True(E.IsLeft(result), "expected Left for missing config path")

	err := F.Pipe1(
		result,
		E.Fold(F.Identity[error], func(ResolvedInput) error { return nil }),
	)
	require.NotNil(err)
	require.EqualError(err, "config path is required")
}

func TestValidateInput_MissingDataPath(t *testing.T) {
	require := require.New(t)
	input := Input{
		ConfigPath: "/tmp/config",
		DataPath:   "",
	}

	result := ValidateInput(input)

	require.True(E.IsLeft(result), "expected Left for missing data path")

	err := F.Pipe1(
		result,
		E.Fold(F.Identity[error], func(ResolvedInput) error { return nil }),
	)
	require.NotNil(err)
	require.EqualError(err, "data path is required")
}

func TestValidateInput_ValidMinimalInput(t *testing.T) {
	require := require.New(t)
	input := Input{
		ConfigPath: "/tmp/config",
		DataPath:   "/tmp/data",
	}

	result := ValidateInput(input)

	require.True(E.IsRight(result), "expected Right for valid minimal input")

	resolved := F.Pipe1(
		result,
		E.Fold(func(error) ResolvedInput { return ResolvedInput{} }, F.Identity[ResolvedInput]),
	)
	require.NotEmpty(resolved.ContainerName)
	require.NotEmpty(resolved.ImageName)
	require.Len(resolved.Mounts, 2)
}

func TestValidateInput_InvalidContainerName(t *testing.T) {
	require := require.New(t)
	input := Input{
		ConfigPath:    "/tmp/config",
		DataPath:      "/tmp/data",
		ContainerName: "123invalid",
	}

	result := ValidateInput(input)

	require.True(E.IsLeft(result), "expected Left for invalid container name")

	err := F.Pipe1(
		result,
		E.Fold(F.Identity[error], func(ResolvedInput) error { return nil }),
	)
	require.NotNil(err)
	require.Contains(err.Error(), "container name must start with a letter")
}

func TestValidateInput_ValidContainerName(t *testing.T) {
	require := require.New(t)
	input := Input{
		ConfigPath:    "/tmp/config",
		DataPath:      "/tmp/data",
		ContainerName: "my-container_123",
	}

	result := ValidateInput(input)

	require.True(E.IsRight(result), "expected Right for valid container name")

	resolved := F.Pipe1(
		result,
		E.Fold(func(error) ResolvedInput { return ResolvedInput{} }, F.Identity[ResolvedInput]),
	)
	require.Equal("my-container_123", resolved.ContainerName)
}

func TestValidateInput_ReservedVolumeBasename(t *testing.T) {
	require := require.New(t)
	input := Input{
		ConfigPath: "/tmp/config",
		DataPath:   "/tmp/data",
		Volumes:    []string{"/tmp/config"},
	}

	result := ValidateInput(input)

	require.True(E.IsLeft(result), "expected Left for reserved volume basename")

	err := F.Pipe1(
		result,
		E.Fold(F.Identity[error], func(ResolvedInput) error { return nil }),
	)
	require.NotNil(err)
	require.Contains(err.Error(), "reserved or invalid")
}

func TestValidateInput_ValidVolume(t *testing.T) {
	require := require.New(t)
	input := Input{
		ConfigPath: "/tmp/config",
		DataPath:   "/tmp/data",
		Volumes:    []string{"/tmp/myproject"},
	}

	result := ValidateInput(input)

	require.True(E.IsRight(result), "expected Right for valid volume")

	resolved := F.Pipe1(
		result,
		E.Fold(func(error) ResolvedInput { return ResolvedInput{} }, F.Identity[ResolvedInput]),
	)
	require.Len(resolved.Mounts, 3)
}

func TestValidateInput_WorkspaceOptional(t *testing.T) {
	require := require.New(t)
	input := Input{
		ConfigPath:    "/tmp/config",
		DataPath:      "/tmp/data",
		WorkspacePath: "/tmp/workspace",
	}

	result := ValidateInput(input)

	require.True(E.IsRight(result), "expected Right with workspace")

	resolved := F.Pipe1(
		result,
		E.Fold(func(error) ResolvedInput { return ResolvedInput{} }, F.Identity[ResolvedInput]),
	)
	require.NotEmpty(resolved.Workdir)
	require.Len(resolved.Mounts, 3)
}

func TestValidateInput_DefaultImageName(t *testing.T) {
	require := require.New(t)
	input := Input{
		ConfigPath: "/tmp/config",
		DataPath:   "/tmp/data",
		ImageName:  "",
	}

	result := ValidateInput(input)

	require.True(E.IsRight(result), "expected Right with default image name")

	resolved := F.Pipe1(
		result,
		E.Fold(func(error) ResolvedInput { return ResolvedInput{} }, F.Identity[ResolvedInput]),
	)
	require.Equal(DefaultImageName, resolved.ImageName)
}

func TestValidateInput_CustomImageName(t *testing.T) {
	require := require.New(t)
	input := Input{
		ConfigPath: "/tmp/config",
		DataPath:   "/tmp/data",
		ImageName:  "custom:v1.0",
	}

	result := ValidateInput(input)

	require.True(E.IsRight(result), "expected Right with custom image name")

	resolved := F.Pipe1(
		result,
		E.Fold(func(error) ResolvedInput { return ResolvedInput{} }, F.Identity[ResolvedInput]),
	)
	require.Equal("custom:v1.0", resolved.ImageName)
}

func TestIsReservedBasename(t *testing.T) {
	tests := []struct {
		name     string
		basename string
		expected bool
	}{
		{"config is reserved", "config", true},
		{"data is reserved", "data", true},
		{"crush is reserved", "crush", true},
		{"project is not reserved", "project", false},
		{"workspace is not reserved", "workspace", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			require.Equal(tt.expected, isReservedBasename(tt.basename))
		})
	}
}
