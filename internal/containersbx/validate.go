package containersbx

import (
	"errors"
	"io/fs"
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
	ORD "github.com/IBM/fp-go/v2/ord"
	Pred "github.com/IBM/fp-go/v2/predicate"
	Str "github.com/IBM/fp-go/v2/string"

	FP "github.com/cybersiddhu/crush-sandbox/internal/fp"
)

const (
	alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	charNo   = 6
)

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
		F.Constant1[os.FileInfo](errors.New("skills path is not a directory")),
	)

	intOrd            = ORD.FromStrictCompare[int]()
	isGreaterThanZero = ORD.Gt(intOrd)(0)
	isNonEmptyDir     = E.FromPredicate(
		Pred.ContraMap(A.Size[fs.DirEntry])(isGreaterThanZero),
		F.Constant1[[]fs.DirEntry](errors.New("skills directory is empty")),
	)

	readDir = IOE.Eitherize1(os.ReadDir)
)

func randomByte() I.IO[byte] {
	return F.Pipe1(
		randAlpha,
		I.Map(func(i int) byte { return alphabet[i] }),
	)
}

// generateKitName creates a random kit name like "crush-sbx-aB3xZ".
func generateKitName(n int) I.IO[string] {
	return F.Pipe3(
		A.Replicate(n, randomByte()),
		I.SequenceArray[byte],
		I.Map(func(bs []byte) string {
			return string(bs)
		}),
		I.Map(Str.Prepend("crush-sbx")),
	)
}

// NormalizeInput fills default values for blank fields.
func NormalizeInput(input Input) Input {
	return Input{
		ConfigPath:   input.ConfigPath,
		SkillsPath:   input.SkillsPath,
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
	return F.Pipe3(
		input.APIKey,
		O.FromPredicate(Str.IsNonEmpty),
		E.FromOption[string](Err.OnNone("API key must not be empty")),
		E.MapTo[error, string](input),
	)
}

func validateConfigPath(input Input) E.Either[error, Input] {
	return F.Pipe2(
		input.ConfigPath,
		O.FromPredicate(Str.IsNonEmpty),
		O.Fold(
			func() E.Either[error, Input] { return E.Of[error](input) },
			func(path string) E.Either[error, Input] {
				return F.Pipe3(
					IOE.Of[error](path),
					IOE.Chain(FILE.Stat),
					IOE.Map[error](F.Constant1[os.FileInfo](input)),
					FP.ToEither[error, Input],
				)
			},
		),
	)
}

func validateSkillsPath(input Input) E.Either[error, Input] {
	return F.Pipe2(
		input.SkillsPath,
		O.FromPredicate(Str.IsNonEmpty),
		O.Fold(
			func() E.Either[error, Input] { return E.Of[error](input) },
			func(path string) E.Either[error, Input] {
				return F.Pipe2(
					validateNonEmptyDir(path),
					IOE.Map[error](F.Constant1[string](input)),
					FP.ToEither[error, Input],
				)
			},
		),
	)
}

func validateNonEmptyDir(path string) IOE.IOEither[error, string] {
	return F.Pipe4(
		FILE.Stat(path),
		IOE.ChainEitherK(isDirectory),
		IOE.Chain(F.Constant1[os.FileInfo](readDir(path))),
		IOE.ChainEitherK(isNonEmptyDir),
		IOE.Map[error](F.Constant1[[]fs.DirEntry](path)),
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
