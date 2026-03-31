// Package fp provides functional programming utilities for Go, including
// conversions between different functional types.
package fp

import (
	E "github.com/IBM/fp-go/v2/either"
	IOE "github.com/IBM/fp-go/v2/ioeither"
)

// ToEither evaluates an IOEither and returns the underlying Either.
// Since IOEither is a lazy computation, this forces evaluation.
func ToEither[Err, A any](ioe IOE.IOEither[Err, A]) E.Either[Err, A] {
	return ioe()
}
