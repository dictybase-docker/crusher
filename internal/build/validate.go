package build

import (
	"errors"
	"strings"

	A "github.com/IBM/fp-go/v2/array"
	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	Str "github.com/IBM/fp-go/v2/string"
)

var nonEmpty = F.Flow2(strings.TrimSpace, Str.IsNonEmpty)

func allNonEmpty(values []string) bool {
	return F.Pipe1(
		values,
		A.Reduce(func(acc bool, item string) bool {
			return acc && nonEmpty(item)
		}, true),
	)
}

func ValidateRequest(r Request) E.Either[error, Request] {
	validations := []E.Either[error, bool]{
		F.Pipe2(
			r.File,
			E.FromPredicate(
				nonEmpty,
				func(string) error { return errors.New("dockerfile path is required") },
			),
			E.MapTo[error, string](true),
		),
		F.Pipe2(
			r.Tags,
			E.FromPredicate(
				A.IsNonEmpty[string],
				func([]string) error { return errors.New("at least one tag is required") },
			),
			E.MapTo[error, []string](true),
		),
		F.Pipe2(
			r.Tags,
			E.FromPredicate(
				allNonEmpty,
				func([]string) error { return errors.New("tag values must be non-empty") },
			),
			E.MapTo[error, []string](true),
		),
	}

	return F.Pipe1(
		E.SequenceArray(validations),
		E.MapTo[error, []bool](r),
	)
}
