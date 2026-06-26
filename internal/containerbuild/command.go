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
	FP "github.com/dictybase-docker/crusher/internal/fp"
	"github.com/urfave/cli/v3"
)

var (
	// lookupResolver is a curried function that selects the Dockerfile resolver
	// from the Record by embed flag.
	lookupResolver = F.Curry2(
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
		Name:  buildCmd,
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
				Value:   []string{latestTag},
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
				Value: defaultGolangciLintVersion,
			},
			&cli.StringFlag{
				Name:  "crush-version",
				Usage: "crush version",
				Value: latestTag,
			},
			&cli.StringFlag{
				Name:  "gotestsum-version",
				Usage: "gotestsum version",
				Value: latestTag,
			},
			&cli.StringFlag{
				Name:  "moxide-version",
				Usage: "markdown-oxide version",
				Value: latestTag,
			},
			&cli.StringFlag{
				Name:  "sem-version",
				Usage: "sem version",
				Value: latestTag,
			},
			&cli.StringFlag{
				Name:  "rtk-version",
				Usage: "rtk version",
				Value: latestTag,
			},
		},
		Action: Action,
	}
}

// Action is the build subcommand entry point.
// Pipeline: validate tags → acquire dockerfile → render args → run process → release.
// Folds into Pair; printing (a no-op for build) is outside the pipeline.
func Action(ctx context.Context, cmd *cli.Command) error {
	result := F.Pipe5(
		InputFromCommand(ctx, cmd),
		ValidateInput,
		IOE.FromEither[error],
		IOE.Chain(Execute),
		FP.ToEither[error, F.Void],
		E.Fold(
			func(err error) P.Pair[error, F.Void] {
				return P.MakePair(err, F.VOID)
			},
			func(v F.Void) P.Pair[error, F.Void] {
				return P.MakePair[error](nil, v)
			},
		),
	)

	if err := P.Head(result); err != nil {
		return err
	}

	F.Pipe2(result, P.Tail, printResult)

	return nil
}

// printResult is a no-op for the build subcommand: the build pipeline
// produces no user-facing result to surface. It exists to keep the Action
// shape consistent with the other subcommands.
func printResult(F.Void) F.Void {
	return F.VOID
}
