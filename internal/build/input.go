package build

import (
	"context"

	IOE "github.com/IBM/fp-go/v2/ioeither"
)

const containerBinary = "container"

// Input holds the build parameters throughout the pipeline.
// DockerfileSource is a lazy IOEither that resolves to a DockerfileResource
// when executed. It is set once in InputFromCommand and never branched on.
type Input struct {
	DockerfileSource IOE.IOEither[error, DockerfileResource]
	Name             string
	Tags             []string
	Ctx              context.Context
}

// CommandSpec holds the resolved executable binary and argv slice.
// Built inside Execute after the DockerfileResource is acquired.
type CommandSpec struct {
	Bin  string
	Args []string
}
