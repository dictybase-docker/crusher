package containerbuild

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

	// isNonBlank is the negation of isBlank.
	isNonBlank = P.Not(isBlank)

	// A tag list is valid when it has at least one item and none of them
	// are blank after trimming whitespace.
	isAllNonBlank = F.Pipe1(
		A.IsNonEmpty[string],
		P.And(P.Not(A.Any(isBlank))),
	)
)

// ValidateInput checks that the tags in the Input are valid according to
func ValidateInput(r Input) E.Either[error, Input] {
	return F.Pipe2(
		r.Tags,
		E.FromPredicate(
			isAllNonBlank,
			func([]string) error {
				return errors.New("tag values must be non-empty")
			},
		),
		E.MapTo[error, []string](r),
	)
}
