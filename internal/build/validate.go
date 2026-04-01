package build

import (
	"errors"
	"strings"

	A "github.com/IBM/fp-go/v2/array"
	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	P "github.com/IBM/fp-go/v2/predicate"
	Str "github.com/IBM/fp-go/v2/string"
)

var (
	// isBlank is true when the input becomes empty after trimming whitespace.
	isBlank = F.Pipe1(
		Str.IsEmpty,
		P.ContraMap(strings.TrimSpace),
	)

	// A tag list is valid when it has at least one item and none of them
	// are blank after trimming whitespace.
	isAllNonBlank = F.Pipe1(
		A.IsNonEmpty[string],
		P.And(P.Not(A.Any(isBlank))),
	)
)

func ValidateRequest(r Request) E.Either[error, Request] {
	validations := []E.Either[error, bool]{
		F.Pipe2(
			r.File,
			E.FromPredicate(
				P.Not(isBlank),
				func(string) error {
					return errors.New("dockerfile path is required")
				},
			),
			E.MapTo[error, string](true),
		),
		F.Pipe2(
			r.Tags,
			E.FromPredicate(
				isAllNonBlank,
				func([]string) error {
					return errors.New("tag values must be non-empty")
				},
			),
			E.MapTo[error, []string](true),
		),
	}

	return F.Pipe1(
		E.SequenceArray(validations),
		E.MapTo[error, []bool](r),
	)
}
