// Package containerbuild implements the "build" subcommand, which builds an OCI
// image using the docker CLI.
package containerbuild

import (
	"context"

	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	O "github.com/IBM/fp-go/v2/option"
	P "github.com/IBM/fp-go/v2/pair"
	R "github.com/IBM/fp-go/v2/record"
	FP "github.com/cybersiddhu/crush-sandbox/internal/fp"
	"github.com/urfave/cli/v3"
)

// lookupResolver is a curried function that selects the Dockerfile resolver
// from the Record by embed flag.
var lookupResolver = F.Curry2(
	func(embed bool, record map[bool]IOE.IOEither[error, DockerfileResource]) IOE.IOEither[error, DockerfileResource] {
		return F.Pipe2(
			record,
			R.Lookup[IOE.IOEither[error, DockerfileResource]](embed),
			O.GetOrElse(func() IOE.IOEither[error, DockerfileResource] {
				return EmbeddedResolver()
			}),
		)
	},
)

// resolverEntries builds a Record of the two Dockerfile resolver strategies
// keyed by the --embed flag.
func resolverEntries(cmd *cli.Command) R.Entries[bool, IOE.IOEither[error, DockerfileResource]] {
	return R.Entries[bool, IOE.IOEither[error, DockerfileResource]]{
		P.MakePair(false, FileResolver(cmd.String("file"))),
		P.MakePair(true, EmbeddedResolver()),
	}
}

// buildArgEntries reads the versions from the CLI and constructs a record
func buildArgEntries(cmd *cli.Command) R.Entries[string, string] {
	return R.Entries[string, string]{
		P.MakePair("GOLANGCI_LINT_VERSION", cmd.String("golangci-lint-version")),
		P.MakePair("CRUSH_VERSION", cmd.String("crush-version")),
		P.MakePair("GOTESTSUM_VERSION", cmd.String("gotestsum-version")),
		P.MakePair("MOXIDE_VERSION", cmd.String("moxide-version")),
		P.MakePair("SEM_VERSION", cmd.String("sem-version")),
		P.MakePair("RTK_VERSION", cmd.String("rtk-version")),
	}
}

// InputFromCommand reads CLI flags and constructs the build Input
// using Record-based dispatch for both Dockerfile source and build args.
func InputFromCommand(ctx context.Context, cmd *cli.Command) Input {
	return Input{
		Ctx:       ctx,
		Name:      cmd.String("name"),
		Tags:      cmd.StringSlice("tag"),
		BuildArgs: F.Pipe2(cmd, buildArgEntries, R.FromEntries),
		DockerfileSource: F.Pipe3(
			cmd,
			resolverEntries,
			R.FromEntries,
			lookupResolver(cmd.Bool("embed")),
		),
	}
}

func Command() *cli.Command {
	return &cli.Command{
		Name:  "build",
		Usage: "Build an OCI image via the docker CLI",
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
			&cli.StringFlag{
				Name:  "moxide-version",
				Usage: "markdown-oxide version",
				Value: "latest",
			},
			&cli.StringFlag{
				Name:  "sem-version",
				Usage: "sem version",
				Value: "latest",
			},
			&cli.StringFlag{
				Name:  "rtk-version",
				Usage: "rtk version",
				Value: "latest",
			},
		},
		Action: Action,
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
