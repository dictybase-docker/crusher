// Package predord provides reusable fp-go ord/eq instances and derived
// predicates.
package predord

import (
	"strings"

	EQ "github.com/IBM/fp-go/v2/eq"
	F "github.com/IBM/fp-go/v2/function"
	ORD "github.com/IBM/fp-go/v2/ord"
	Pred "github.com/IBM/fp-go/v2/predicate"
	Str "github.com/IBM/fp-go/v2/string"
)

// Reusable fp-go ord/eq instances and derived predicates.
var (
	IntOrd     = ORD.FromStrictCompare[int]()
	Float64Ord = ORD.FromStrictCompare[float64]()
	IntEq      = EQ.FromStrictEquals[int]()
	Float64Eq  = EQ.FromStrictEquals[float64]()
	StringEq   = EQ.FromStrictEquals[string]()

	IntBetween = ORD.Between(IntOrd)

	IntBetweenInclusive = func(lo, hi int) Pred.Predicate[int] {
		return Pred.And(ORD.Geq(IntOrd)(lo))(ORD.Leq(IntOrd)(hi))
	}

	MinStrLen = F.Flow2(
		ORD.Geq(IntOrd),
		Pred.ContraMap(Str.Size),
	)

	MaxStrLen = F.Flow2(
		ORD.Leq(IntOrd),
		Pred.ContraMap(Str.Size),
	)

	StrLenEq = F.Flow2(
		EQ.Equals(IntEq),
		Pred.ContraMap(Str.Size),
	)

	NotEqualF64 = F.Flow2(EQ.Equals(Float64Eq), Pred.Not)
	NotEqualInt = F.Flow2(EQ.Equals(IntEq), Pred.Not)
	NotEqualStr = F.Flow2(EQ.Equals(StringEq), Pred.Not)
	StrEq       = EQ.Equals(StringEq)

	// IsBlank is true when a string becomes empty after trimming whitespace.
	IsBlank = F.Pipe1(
		Str.IsEmpty,
		Pred.ContraMap(strings.TrimSpace),
	)

	// IsNonBlank is the negation of IsBlank.
	IsNonBlank = Pred.Not(IsBlank)
)