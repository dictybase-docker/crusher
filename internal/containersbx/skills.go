package containersbx

import (
	"os"
	"path/filepath"

	F "github.com/IBM/fp-go/v2/function"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	O "github.com/IBM/fp-go/v2/option"
)

// ReadSkills reads the skills directory tree into a map of skill name to SKILL.md content.
// Blank skillsPath returns an empty map.
func ReadSkills(skillsPath string) IOE.IOEither[error, map[string]string] {
	return F.Pipe2(
		skillsPath,
		O.FromPredicate(func(s string) bool { return s != "" }),
		O.Fold(
			func() IOE.IOEither[error, map[string]string] {
				return IOE.Of[error](map[string]string{})
			},
			func(path string) IOE.IOEither[error, map[string]string] {
				return IOE.TryCatchError(func() (map[string]string, error) {
					entries, err := os.ReadDir(path)
					if err != nil {
						return nil, err
					}
					result := make(map[string]string)
					for _, entry := range entries {
						if !entry.IsDir() {
							continue
						}
						skillFile := filepath.Join(path, entry.Name(), "SKILL.md")
						data, err := os.ReadFile(skillFile)
						if err != nil {
							continue
						}
						result[entry.Name()] = string(data)
					}
					return result, nil
				})
			},
		),
	)
}
