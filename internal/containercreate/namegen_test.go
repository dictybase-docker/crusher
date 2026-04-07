package containercreate

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenerateName_Format(t *testing.T) {
	require := require.New(t)

	name := GenerateName()

	require.NotEmpty(name, "name should not be empty")
	require.Contains(name, "-", "name should contain a hyphen")

	parts := strings.Split(name, "-")
	require.Len(parts, 2, "name should have exactly two parts separated by hyphen")
	require.NotEmpty(parts[0], "adjective should not be empty")
	require.NotEmpty(parts[1], "noun should not be empty")
}

func TestGenerateName_Unique(t *testing.T) {
	require := require.New(t)

	names := make(map[string]bool)
	for i := 0; i < 100; i++ {
		name := GenerateName()
		names[name] = true
	}

	require.Greater(len(names), 50, "should generate diverse names")
}

func TestGenerateName_StartsWithLetter(t *testing.T) {
	require := require.New(t)

	for i := 0; i < 100; i++ {
		name := GenerateName()
		require.True(isValidContainerName(name), "generated name should be a valid container name")
	}
}

func TestGenerateName_MatchesPattern(t *testing.T) {
	require := require.New(t)

	for i := 0; i < 100; i++ {
		name := GenerateName()
		require.Regexp(`^[a-z]+-[a-z]+$`, name, "name should match adjective-noun pattern")
	}
}

func TestPickRandom_ReturnsValueFromSlice(t *testing.T) {
	require := require.New(t)

	items := []string{"a", "b", "c"}
	for i := 0; i < 100; i++ {
		result := pickRandom(items)
		require.Contains(items, result, "result should be from the slice")
	}
}

func TestPickRandom_EmptySlice(t *testing.T) {
	require := require.New(t)

	var items []string
	result := pickRandom(items)
	require.Equal("", result, "empty slice should return zero value")
}

func TestPickRandom_SingleElement(t *testing.T) {
	require := require.New(t)

	items := []string{"only"}
	for i := 0; i < 10; i++ {
		result := pickRandom(items)
		require.Equal("only", result, "single element slice should return that element")
	}
}
