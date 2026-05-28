package containersbx

import (
	"os"
	"strings"

	IOE "github.com/IBM/fp-go/v2/ioeither"
	O "github.com/IBM/fp-go/v2/option"
)

// ReadConfig reads the user's crush.json or returns the default OpenRouter config.
func ReadConfig(configPath string) IOE.IOEither[error, string] {
	return O.Fold(
		func() IOE.IOEither[error, string] {
			return IOE.Of[error](DefaultConfig())
		},
		func(path string) IOE.IOEither[error, string] {
			return IOE.TryCatchError(func() (string, error) {
				data, err := os.ReadFile(path)
				if err != nil {
					return "", err
				}
				return string(data), nil
			})
		},
	)(O.FromPredicate(func(s string) bool {
		return strings.TrimSpace(s) != ""
	})(configPath))
}

// DefaultConfig returns the default OpenRouter-only crush.json.
func DefaultConfig() string {
	return `{
  "model": "openai/gpt-4o",
  "providers": {
    "openrouter": {
      "id": "openrouter",
      "name": "OpenRouter",
      "base_url": "https://openrouter.ai/api/v1",
      "type": "openai"
    }
  },
  "options": {
    "permissions": {
      "mode": "default"
    }
  }
}`
}
