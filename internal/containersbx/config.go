package containersbx

import (
	F "github.com/IBM/fp-go/v2/function"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	FILE "github.com/IBM/fp-go/v2/ioeither/file"
	O "github.com/IBM/fp-go/v2/option"
)

// ReadConfig reads the user's crush.json or returns the default OpenRouter config.
func ReadConfig(enriched Input) IOE.IOEither[error, genState] {
	return F.Pipe3(
		enriched.ConfigPath,
		// string -> Option[string]  (None if blank)
		O.FromPredicate(isNonBlank),
		// Option[string] -> IOEither[error, string]
		O.Fold(
			func() IOE.IOEither[error, string] {
				return IOE.Of[error](DefaultConfig())
			},
			func(path string) IOE.IOEither[error, string] {
				return F.Pipe1(
					FILE.ReadFile(path),
					IOE.Map[error](func(bs []byte) string {
						return string(bs)
					}),
				)
			},
		),
		// string -> genState
		IOE.Map[error](func(configContent string) genState {
			return genState{
				input:         enriched,
				configContent: configContent,
			}
		}),
	)
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
