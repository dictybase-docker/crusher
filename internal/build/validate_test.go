package build

import (
	"testing"

	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	"github.com/stretchr/testify/require"
)

func TestValidateInput_EmptyFile(t *testing.T) {
	require := require.New(t)
	req := Input{
		File: "",
		Tags: []string{"latest"},
	}

	result := ValidateInput(req)

	require.True(E.IsLeft(result), "expected Left for empty Dockerfile path")

	err := F.Pipe1(
		result,
		E.Fold(F.Identity[error], func(Input) error { return nil }),
	)
	require.NotNil(err)
	require.EqualError(err, "dockerfile path is required")
}

func TestValidateInput_EmptyTagList(t *testing.T) {
	require := require.New(t)
	req := Input{
		File: "Dockerfile",
		Tags: []string{},
	}

	result := ValidateInput(req)

	require.True(E.IsLeft(result), "expected Left for empty tag list")

	err := F.Pipe1(
		result,
		E.Fold(F.Identity[error], func(Input) error { return nil }),
	)
	require.NotNil(err)
	require.EqualError(err, "tag values must be non-empty")
}

func TestValidateInput_BlankTagEntry(t *testing.T) {
	require := require.New(t)
	req := Input{
		File: "Dockerfile",
		Tags: []string{"latest", ""},
	}

	result := ValidateInput(req)

	require.True(E.IsLeft(result), "expected Left for blank tag entry")

	err := F.Pipe1(
		result,
		E.Fold(F.Identity[error], func(Input) error { return nil }),
	)
	require.NotNil(err)
	require.EqualError(err, "tag values must be non-empty")
}

func TestValidateInput_DefaultInput(t *testing.T) {
	require := require.New(t)
	req := Input{
		File: "Dockerfile",
		Tags: []string{"latest"},
	}

	result := ValidateInput(req)

	require.True(E.IsRight(result), "expected Right for default input")

	validated := F.Pipe1(
		result,
		E.Fold(func(error) Input { return Input{} }, F.Identity[Input]),
	)
	require.Equal("Dockerfile", validated.File)
	require.Equal([]string{"latest"}, validated.Tags)
}

func TestValidateInput_MultipleTags(t *testing.T) {
	require := require.New(t)
	req := Input{
		File: "docker/Prod.Dockerfile",
		Tags: []string{"latest", "stable", "v1.0.0"},
	}

	result := ValidateInput(req)

	require.True(E.IsRight(result), "expected Right for valid build input with multiple tags")

	validated := F.Pipe1(
		result,
		E.Fold(func(error) Input { return Input{} }, F.Identity[Input]),
	)
	require.Equal("docker/Prod.Dockerfile", validated.File)
	require.Equal([]string{"latest", "stable", "v1.0.0"}, validated.Tags)
}
