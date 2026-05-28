package containersbx

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	O "github.com/IBM/fp-go/v2/option"
	Pred "github.com/IBM/fp-go/v2/predicate"
	Str "github.com/IBM/fp-go/v2/string"
)

var (
	isBlank = F.Pipe1(
		Str.IsEmpty,
		Pred.ContraMap(strings.TrimSpace),
	)

	isNonBlank = Pred.Not(isBlank)
)

// NormalizeInput fills default values for blank fields.
func NormalizeInput(input Input) Input {
	return Input{
		OutputPath: F.Pipe2(
			input.OutputPath,
			O.FromPredicate(isNonBlank),
			O.GetOrElse(func() string { return DefaultOutputPath }),
		),
		KitName: F.Pipe2(
			input.KitName,
			O.FromPredicate(isNonBlank),
			O.GetOrElse(func() string {
				cwd, err := os.Getwd()
				if err != nil {
					return "crush-sbx"
				}
				return filepath.Base(cwd)
			}),
		),
		ConfigPath:   input.ConfigPath,
		SkillsPath:   input.SkillsPath,
		APIKey:       input.APIKey,
		ShouldCreate: input.ShouldCreate,
		CrushVersion: F.Pipe2(
			input.CrushVersion,
			O.FromPredicate(isNonBlank),
			O.GetOrElse(func() string { return DefaultCrushVersion }),
		),
		GolangciLintVersion: F.Pipe2(
			input.GolangciLintVersion,
			O.FromPredicate(isNonBlank),
			O.GetOrElse(func() string { return DefaultGolangciLintVersion }),
		),
		GoVersion: F.Pipe2(
			input.GoVersion,
			O.FromPredicate(isNonBlank),
			O.GetOrElse(func() string { return DefaultGoVersion }),
		),
		GotestsumVersion: F.Pipe2(
			input.GotestsumVersion,
			O.FromPredicate(isNonBlank),
			O.GetOrElse(func() string { return DefaultGotestsumVersion }),
		),
		MoxideVersion: F.Pipe2(
			input.MoxideVersion,
			O.FromPredicate(isNonBlank),
			O.GetOrElse(func() string { return DefaultMoxideVersion }),
		),
		SemVersion: F.Pipe2(
			input.SemVersion,
			O.FromPredicate(isNonBlank),
			O.GetOrElse(func() string { return DefaultSemVersion }),
		),
		RtkVersion: F.Pipe2(
			input.RtkVersion,
			O.FromPredicate(isNonBlank),
			O.GetOrElse(func() string { return DefaultRtkVersion }),
		),
		Ctx: input.Ctx,
	}
}

// ValidateInput normalizes defaults, validates required fields and paths.
func ValidateInput(input Input) E.Either[error, Input] {
	return F.Pipe5(
		E.Of[error](input),
		E.Map[error](NormalizeInput),
		E.Chain(validateAPIKey),
		E.Chain(validateConfigPath),
		E.Chain(validateSkillsPath),
		E.Chain(validateOutputParent),
	)
}

func validateAPIKey(input Input) E.Either[error, Input] {
	return F.Pipe2(
		input.APIKey,
		E.FromPredicate(
			isNonBlank,
			func(string) error {
				return errors.New("--api-key is required")
			},
		),
		E.MapTo[error, string](input),
	)
}

func validateConfigPath(input Input) E.Either[error, Input] {
	if input.ConfigPath == "" {
		return E.Of[error](input)
	}
	return F.Pipe2(
		input.ConfigPath,
		E.FromPredicate(
			func(p string) bool {
				_, err := os.Stat(p)
				return err == nil
			},
			func(p string) error {
				return errors.New("config file not found: " + p)
			},
		),
		E.MapTo[error, string](input),
	)
}

func validateSkillsPath(input Input) E.Either[error, Input] {
	if input.SkillsPath == "" {
		return E.Of[error](input)
	}
	return F.Pipe2(
		input.SkillsPath,
		E.FromPredicate(
			func(p string) bool {
				info, err := os.Stat(p)
				return err == nil && info.IsDir()
			},
			func(p string) error {
				return errors.New("skills directory not found: " + p)
			},
		),
		E.MapTo[error, string](input),
	)
}

func validateOutputParent(input Input) E.Either[error, Input] {
	return F.Pipe2(
		input.OutputPath,
		E.FromPredicate(
			func(p string) bool {
				parent := filepath.Dir(p)
				info, err := os.Stat(parent)
				return err == nil && info.IsDir()
			},
			func(p string) error {
				return errors.New("output path parent directory does not exist: " + filepath.Dir(p))
			},
		),
		E.MapTo[error, string](input),
	)
}
