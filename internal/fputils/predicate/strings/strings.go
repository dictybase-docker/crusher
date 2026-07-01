// Package predstrings provides predicates on strings, derived from curried
// stdlib string functions.
package predstrings

import (
	"strings"

	F "github.com/IBM/fp-go/v2/function"
	ORD "github.com/IBM/fp-go/v2/ord"
	Pred "github.com/IBM/fp-go/v2/predicate"
	Str "github.com/IBM/fp-go/v2/string"

	predord "github.com/dictybase-docker/crusher/internal/fputils/predicate/ord"
)

var (
	LastIndexOf       = F.Bind2of2(strings.LastIndex)
	HasSuffix         = F.Bind2of2(strings.HasSuffix)
	ContainsRuneClass = F.Bind2of2(strings.ContainsFunc)

	HasAtSign = F.Pipe2(
		0,
		ORD.Geq(predord.IntOrd),
		Pred.ContraMap(LastIndexOf("@")),
	)
)

// StrLenBetween whether len(s) is between first and second (inclusive)
func StrLenBetween(first, second int) Pred.Predicate[string] {
	return Pred.ContraMap(
		Str.Size,
	)(
		predord.IntBetweenInclusive(first, second))
}