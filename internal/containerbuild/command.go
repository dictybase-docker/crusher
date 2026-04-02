// Package build provides a CLI command to build an OCI image using the
// container CLI .
package containerbuild

import (
	"context"

	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	FP "github.com/cybersiddhu/crush-sandbox/internal/fp"
	"github.com/urfave/cli/v3"
)

// resolverFactories is a map-based dispatch table that selects the Dockerfile
// resolver strategy based on the --embed flag. This is the ONLY location
// where the boolean is observed — no if/else anywhere in application code.
var resolverFactories = map[bool]func(*cli.Command) IOE.IOEither[error, DockerfileResource]{
	false: func(cmd *cli.Command) IOE.IOEither[error, DockerfileResource] {
		return FileResolver(cmd.String("file"))
	},
	true: func(_ *cli.Command) IOE.IOEither[error, DockerfileResource] {
		return EmbeddedResolver()
	},
}

func Command() *cli.Command {
	return &cli.Command{
		Name:  "build",
		Usage: "Build an OCI image via the container CLI",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "file",
				Aliases: []string{"f"},
				Usage:   "Path to Dockerfile",
				Value:   "Dockerfile",
			},
			&cli.StringSliceFlag{
				Name:    "tag",
				Aliases: []string{"t"},
				Usage:   "Image tag, repeatable",
				Value:   []string{"latest"},
			},
			&cli.StringFlag{
				Name:    "name",
				Aliases: []string{"n"},
				Usage:   "Image name (combines with tags as name:tag)",
				Value:   "crusher",
			},
			&cli.BoolFlag{
				Name:  "embed",
				Usage: "Use the Dockerfile embedded in the binary (ignores --file)",
			},
		},
		Action: Action,
	}
}

// InputFromCommand reads CLI flags and selects the Dockerfile resolver
// via the map-based dispatch table.
func InputFromCommand(ctx context.Context, cmd *cli.Command) Input {
	return Input{
		DockerfileSource: resolverFactories[cmd.Bool("embed")](cmd),
		Name:             cmd.String("name"),
		Tags:             cmd.StringSlice("tag"),
		Ctx:              ctx,
	}
}

// Action is the build subcommand entry point.
// Pipeline: validate tags → acquire dockerfile → render args → run process → release.
func Action(ctx context.Context, cmd *cli.Command) error {
	return F.Pipe5(
		InputFromCommand(ctx, cmd),
		ValidateInput,
		IOE.FromEither[error],
		IOE.Chain(Execute),
		FP.ToEither[error, F.Void],
		E.Fold(
			F.Identity[error],
			func(F.Void) error { return nil },
		),
	)
}
