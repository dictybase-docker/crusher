// Package predarrays provides generic slice predicates built with fp-go v2 combinators.
package predarrays

import (
	"slices"

	A "github.com/IBM/fp-go/v2/array"
	EQ "github.com/IBM/fp-go/v2/eq"
	F "github.com/IBM/fp-go/v2/function"
	N "github.com/IBM/fp-go/v2/number"
	ORD "github.com/IBM/fp-go/v2/ord"
	Pred "github.com/IBM/fp-go/v2/predicate"

	predord "github.com/dictybase-docker/crusher/internal/fputils/predicate/ord"
)

// IsNonEmpty returns a Predicate[[]E] that is true when len(s) > 0.
func IsNonEmpty[E any]() Pred.Predicate[[]E] {
	return F.Pipe1(N.MoreThan(0), Pred.ContraMap(A.Size[E]))
}

// MinLen returns a Predicate[[]E] that is true when len(s) >= n.
func MinLen[E any](n int) Pred.Predicate[[]E] {
	return F.Pipe2(
		n,
		ORD.Geq(predord.IntOrd),
		Pred.ContraMap(A.Size[E]),
	)
}

// MaxLen returns a Predicate[[]E] that is true when len(s) <= n.
func MaxLen[E any](n int) Pred.Predicate[[]E] {
	return F.Pipe2(
		n,
		ORD.Leq(predord.IntOrd),
		Pred.ContraMap(A.Size[E]),
	)
}

// LenEq returns a Predicate[[]E] that is true when len(s) == n.
func LenEq[E any](n int) Pred.Predicate[[]E] {
	return F.Pipe2(
		n,
		EQ.Equals(predord.IntEq),
		Pred.ContraMap(A.Size[E]),
	)
}

// Contains returns a curried slices.Contains: Contains(value)(slice) → bool.
func Contains[E comparable](value E) Pred.Predicate[[]E] {
	return F.Bind2of2(slices.Contains[[]E, E])(value)
}

// NotContains returns the negation of Contains.
func NotContains[E comparable](value E) Pred.Predicate[[]E] {
	return F.Pipe1(Contains(value), Pred.Not)
}

// MemberOf returns a Predicate[[]E] true when value is equal (per eq) to
// any element of the slice.
func MemberOf[E any](eq EQ.Eq[E]) func(value E) Pred.Predicate[[]E] {
	return func(value E) Pred.Predicate[[]E] {
		return A.Any(EQ.Equals(eq)(value))
	}
}

// NotMemberOf returns the negation of MemberOf.
func NotMemberOf[E any](eq EQ.Eq[E]) func(value E) Pred.Predicate[[]E] {
	return func(value E) Pred.Predicate[[]E] {
		return F.Pipe1(MemberOf(eq)(value), Pred.Not)
	}
}

// Intersect returns the set intersection of two slices: the elements of
// current that are also members (per eq) of other.
func Intersect[E any](eq EQ.Eq[E]) func(other []E) func(current []E) []E {
	return func(other []E) func(current []E) []E {
		return F.Pipe2(
			other,
			F.Flip(MemberOf(eq)),
			A.Filter[E],
		)
	}
}

// Difference returns the set difference of two slices: the elements of
// current that are not members (per eq) of other.
func Difference[E any](eq EQ.Eq[E]) func(other []E) func(current []E) []E {
	return func(other []E) func(current []E) []E {
		return F.Pipe2(
			other,
			F.Flip(NotMemberOf(eq)),
			A.Filter[E],
		)
	}
}
