package fp

import (
	"errors"
	"testing"

	E "github.com/IBM/fp-go/v2/either"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	"github.com/stretchr/testify/require"
)

func TestToEither_Right(t *testing.T) {
	result := ToEither(IOE.Of[error](42))
	require.True(t, E.IsRight(result))

	val := E.Fold(
		func(_ error) int { return -1 },
		func(v int) int { return v },
	)(result)
	require.Equal(t, 42, val)
}

func TestToEither_Left(t *testing.T) {
	result := ToEither(IOE.Left[int](errors.New("boom")))
	require.True(t, E.IsLeft(result))

	err := E.Fold(
		func(e error) error { return e },
		func(int) error { return nil },
	)(result)
	require.EqualError(t, err, "boom")
}
