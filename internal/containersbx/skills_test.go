package containersbx

import (
	"os"
	"path/filepath"
	"testing"

	E "github.com/IBM/fp-go/v2/either"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadSkills_BlankPath(t *testing.T) {
	eitherSkills := ReadSkills("")()
	assert.True(t, E.IsRight(eitherSkills))
	skills, _ := E.Unwrap(eitherSkills)
	assert.Empty(t, skills)
}

func TestReadSkills_ValidDir(t *testing.T) {
	tmpDir := t.TempDir()
	skillDir := filepath.Join(tmpDir, "my-skill")
	require.NoError(t, os.Mkdir(skillDir, 0o755))
	require.NoError(
		t,
		os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("# My Skill\nContent"), 0o600),
	)

	eitherSkills := ReadSkills(tmpDir)()
	assert.True(t, E.IsRight(eitherSkills))
	skills, _ := E.Unwrap(eitherSkills)
	assert.Len(t, skills, 1)
	assert.Equal(t, "# My Skill\nContent", skills["my-skill"])
}

func TestReadSkills_MissingDir(t *testing.T) {
	eitherSkills := ReadSkills("/nonexistent/skills")()
	assert.True(t, E.IsLeft(eitherSkills))
}

func TestReadSkills_SubdirWithoutSKILLmd(t *testing.T) {
	tmpDir := t.TempDir()
	skillDir := filepath.Join(tmpDir, "empty-skill")
	require.NoError(t, os.Mkdir(skillDir, 0o755))

	eitherSkills := ReadSkills(tmpDir)()
	assert.True(t, E.IsRight(eitherSkills))
	skills, _ := E.Unwrap(eitherSkills)
	assert.Empty(t, skills)
}

func TestReadSkills_EmptyDir(t *testing.T) {
	tmpDir := t.TempDir()
	eitherSkills := ReadSkills(tmpDir)()
	assert.True(t, E.IsRight(eitherSkills))
	skills, _ := E.Unwrap(eitherSkills)
	assert.Empty(t, skills)
}

func TestReadSkills_MultipleSkills(t *testing.T) {
	tmpDir := t.TempDir()
	for _, name := range []string{"skill-a", "skill-b"} {
		skillDir := filepath.Join(tmpDir, name)
		require.NoError(t, os.Mkdir(skillDir, 0o755))
		require.NoError(
			t,
			os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("Content for "+name), 0o600),
		)
	}

	eitherSkills := ReadSkills(tmpDir)()
	assert.True(t, E.IsRight(eitherSkills))
	skills, _ := E.Unwrap(eitherSkills)
	assert.Len(t, skills, 2)
	assert.Equal(t, "Content for skill-a", skills["skill-a"])
	assert.Equal(t, "Content for skill-b", skills["skill-b"])
}
