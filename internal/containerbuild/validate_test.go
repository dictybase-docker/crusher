package containerbuild

import (
	"testing"

	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	"github.com/stretchr/testify/require"
)

func TestValidateInput_EmptyTagList(t *testing.T) {
	require := require.New(t)
	req := Input{
		Tags:      []string{},
		BuildArgs: map[string]string{},
	}

	result := ValidateInput(req)

	require.True(E.IsLeft(result), "expected Left for empty tag list")

	err := F.Pipe1(
		result,
		E.Fold(F.Identity[error], func(Input) error { return nil }),
	)
	require.Error(err)
	require.EqualError(err, "tag values must be non-empty")
}

func TestValidateInput_BlankTagEntry(t *testing.T) {
	require := require.New(t)
	req := Input{
		Tags:      []string{defaultTag, ""},
		BuildArgs: map[string]string{},
	}

	result := ValidateInput(req)

	require.True(E.IsLeft(result), "expected Left for blank tag entry")

	err := F.Pipe1(
		result,
		E.Fold(F.Identity[error], func(Input) error { return nil }),
	)
	require.Error(err)
	require.EqualError(err, "tag values must be non-empty")
}

func TestValidateInput_DefaultInput(t *testing.T) {
	require := require.New(t)
	req := Input{
		Name:      testName,
		Tags:      []string{defaultTag},
		BuildArgs: map[string]string{},
	}

	result := ValidateInput(req)

	require.True(E.IsRight(result), "expected Right for default input")

	validated := F.Pipe1(
		result,
		E.Fold(func(error) Input { return Input{} }, F.Identity[Input]),
	)
	require.Equal(testName, validated.Name)
	require.Equal([]string{defaultTag}, validated.Tags)
}

func TestValidateInput_MultipleTags(t *testing.T) {
	require := require.New(t)
	req := Input{
		Name:      testName,
		Tags:      []string{defaultTag, stableTag, v100Tag},
		BuildArgs: map[string]string{},
	}

	result := ValidateInput(req)

	require.True(E.IsRight(result), "expected Right for valid build input with multiple tags")

	validated := F.Pipe1(
		result,
		E.Fold(func(error) Input { return Input{} }, F.Identity[Input]),
	)
	require.Equal(testName, validated.Name)
	require.Equal([]string{defaultTag, stableTag, v100Tag}, validated.Tags)
}

func TestValidateInput_BuildArgsNotValidated(t *testing.T) {
	require := require.New(t)
	req := Input{
		Name:      testName,
		Tags:      []string{defaultTag},
		BuildArgs: map[string]string{glvKey: "invalid-version"},
	}

	result := ValidateInput(req)
	require.True(E.IsRight(result))
}

func TestValidateInput_NilBuildArgs(t *testing.T) {
	require := require.New(t)
	req := Input{
		Name:      testName,
		Tags:      []string{defaultTag},
		BuildArgs: nil,
	}

	result := ValidateInput(req)
	require.True(E.IsRight(result))
}
