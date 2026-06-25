package containeropencodebx

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"

	A "github.com/IBM/fp-go/v2/array"
	E "github.com/IBM/fp-go/v2/either"
	Err "github.com/IBM/fp-go/v2/errors"
	F "github.com/IBM/fp-go/v2/function"
	I "github.com/IBM/fp-go/v2/io"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	FILE "github.com/IBM/fp-go/v2/ioeither/file"
	O "github.com/IBM/fp-go/v2/option"
	Pred "github.com/IBM/fp-go/v2/predicate"
	R "github.com/IBM/fp-go/v2/record"
	Str "github.com/IBM/fp-go/v2/string"

	FP "github.com/dictybase-docker/crusher/internal/fp"
)

const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var (
	isBlank = F.Pipe1(
		Str.IsEmpty,
		Pred.ContraMap(strings.TrimSpace),
	)

	isNonBlank = Pred.Not(isBlank)

	randAlpha = func() int {
		return F.Pipe2(alphabet, Str.Size, rand.Intn)
	}

	isDirectory = E.FromPredicate(
		os.FileInfo.IsDir,
		F.Constant1[os.FileInfo](errors.New("output parent path is not a directory")),
	)
)

func randomByte() I.IO[byte] {
	return F.Pipe1(
		randAlpha,
		I.Map(func(i int) byte { return alphabet[i] }),
	)
}

// generateKitName creates a random kit name like "opencode-sbx-aB3xZ".
func generateKitName(n int) I.IO[string] {
	return F.Pipe3(
		A.Replicate(n, randomByte()),
		I.SequenceArray[byte],
		I.Map(func(bs []byte) string { return string(bs) }),
		I.Map(Str.Prepend(kitNamePrefix)),
	)
}

// NormalizeInput fills default values for blank fields.
func NormalizeInput(input Input) Input {
	return Input{
		APIKey:       input.APIKey,
		ShouldCreate: input.ShouldCreate,
		Ctx:          input.Ctx,
		OutputPath: F.Pipe2(
			input.OutputPath,
			O.FromPredicate(isNonBlank),
			O.GetOrElse(func() string { return DefaultOutputPath }),
		),
		KitName: F.Pipe2(
			input.KitName,
			O.FromPredicate(isNonBlank),
			O.GetOrElse(generateKitName(charNo)),
		),
		AgentImage: F.Pipe2(
			input.AgentImage,
			O.FromPredicate(isNonBlank),
			O.GetOrElse(func() string { return DefaultAgentImage }),
		),
		Provider: F.Pipe2(
			input.Provider,
			O.FromPredicate(isNonBlank),
			O.GetOrElse(func() string { return DefaultProvider }),
		),
		GolangciLintVersion: F.Pipe2(
			input.GolangciLintVersion,
			O.FromPredicate(isNonBlank),
			O.GetOrElse(func() string { return DefaultGolangciLintVersion }),
		),
	}
}

// ValidateInput normalizes defaults, validates required fields and paths.
func ValidateInput(input Input) E.Either[error, Input] {
	return F.Pipe4(
		E.Of[error](input),
		E.Map[error](NormalizeInput),
		E.Chain(validateAPIKey),
		E.Chain(validateProvider),
		E.Chain(validateOutputParent),
	)
}

func validateAPIKey(input Input) E.Either[error, Input] {
	return F.Pipe3(
		input.APIKey,
		O.FromPredicate(Str.IsNonEmpty),
		E.FromOption[string](Err.OnNone("API key must not be empty")),
		E.MapTo[error, string](input),
	)
}

func validateProvider(input Input) E.Either[error, Input] {
	return F.Pipe2(
		R.Lookup[ProviderConfig](input.Provider)(providerConfigs),
		E.FromOption[ProviderConfig](func() error {
			return fmt.Errorf(
				"unsupported provider %q, valid values: openrouter, anthropic, openai, google",
				input.Provider,
			)
		}),
		E.MapTo[error, ProviderConfig](input),
	)
}

func validateOutputParent(input Input) E.Either[error, Input] {
	return F.Pipe2(
		input.OutputPath,
		O.FromPredicate(Str.IsNonEmpty),
		O.Fold(
			func() E.Either[error, Input] { return E.Of[error](input) },
			func(path string) E.Either[error, Input] {
				return F.Pipe4(
					IOE.Of[error](filepath.Dir(path)),
					IOE.Chain(FILE.Stat),
					IOE.ChainEitherK(isDirectory),
					IOE.Map[error](F.Constant1[os.FileInfo](input)),
					FP.ToEither[error, Input],
				)
			},
		),
	)
}
