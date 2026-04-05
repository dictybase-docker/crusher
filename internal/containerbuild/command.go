// Package containerbuild implements the "build" subcommand, which builds an OCI
// image using the container CLI.
package containerbuild

import (
	"context"

	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	P "github.com/IBM/fp-go/v2/pair"
	R "github.com/IBM/fp-go/v2/record"
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

// buildArgEntries reads the versions from the CLI and constructs a record
func buildArgEntries(cmd *cli.Command) R.Entries[string, string] {
	return R.Entries[string, string]{
		P.MakePair("GOLANGCI_LINT_VERSION", cmd.String("golangci-lint-version")),
		P.MakePair("CRUSH_VERSION", cmd.String("crush-version")),
		P.MakePair("GOTESTSUM_VERSION", cmd.String("gotestsum-version")),
	}
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
			&cli.StringFlag{
				Name:  "golangci-lint-version",
				Usage: "golangci-lint version",
				Value: "2.11.4",
			},
			&cli.StringFlag{
				Name:  "crush-version",
				Usage: "crush version",
				Value: "latest",
			},
			&cli.StringFlag{
				Name:  "gotestsum-version",
				Usage: "gotestsum version",
				Value: "latest",
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
		BuildArgs:        F.Pipe2(cmd, buildArgEntries, R.FromEntries),
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
