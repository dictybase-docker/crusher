package containeropencodebx

import (
	F "github.com/IBM/fp-go/v2/function"
	L "github.com/IBM/fp-go/v2/optics/lens"
)

// Lenses for the pipeline's working state. Each setter touches exactly one
// field, so builders update a single field without re-listing the whole
// struct. Composition reaches KitResult leaves from a stepState root.
var (
	// stateLens focuses stepState.State.
	stateLens = L.MakeLens(
		func(s stepState) execState { return s.State },
		func(s stepState, e execState) stepState { s.State = e; return s },
	)

	// resultLens focuses execState.Result.
	resultLens = L.MakeLens(
		func(e execState) KitResult { return e.Result },
		func(e execState, r KitResult) execState { e.Result = r; return e },
	)

	// KitResult leaf lenses.
	kitResultOutputPathLens = L.MakeLens(
		func(r KitResult) string { return r.OutputPath },
		func(r KitResult, v string) KitResult { r.OutputPath = v; return r },
	)
	kitResultKitNameLens = L.MakeLens(
		func(r KitResult) string { return r.KitName },
		func(r KitResult, v string) KitResult { r.KitName = v; return r },
	)
	kitResultCreatedLens = L.MakeLens(
		func(r KitResult) bool { return r.Created },
		func(r KitResult, v bool) KitResult { r.Created = v; return r },
	)

	// Composed lenses rooted at stepState, reaching KitResult leaves.
	ssResult           = F.Pipe1(stateLens, L.Compose[stepState](resultLens))
	ssResultOutputPath = F.Pipe1(ssResult, L.Compose[stepState](kitResultOutputPathLens))
	ssResultKitName    = F.Pipe1(ssResult, L.Compose[stepState](kitResultKitNameLens))
	ssResultCreated    = F.Pipe1(ssResult, L.Compose[stepState](kitResultCreatedLens))
)
