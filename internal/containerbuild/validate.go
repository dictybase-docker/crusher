package containerbuild

import (
	"errors"

	A "github.com/IBM/fp-go/v2/array"
	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	P "github.com/IBM/fp-go/v2/predicate"

	predord "github.com/dictybase-docker/crusher/internal/fputils/predicate/ord"
)

var (
	// A tag list is valid when it has at least one item and none of them
	// are blank after trimming whitespace.
	isAllNonBlank = F.Pipe1(
		A.IsNonEmpty[string],
		P.And(P.Not(A.Any(predord.IsBlank))),
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
