package build

import (
	"context"

	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	"github.com/urfave/cli/v3"
)

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
		},
		Action: Action,
	}
}

func RequestFromCommand(cmd *cli.Command) Request {
	return Request{
		File: cmd.String("file"),
		Tags: cmd.StringSlice("tag"),
	}
}

func Action(ctx context.Context, cmd *cli.Command) error {
	req := RequestFromCommand(cmd)

	program := F.Pipe2(
		IOE.FromEither[error](ValidateRequest(req)),
		IOE.Map[error](RenderCommand),
		IOE.Chain(func(spec CommandSpec) IOE.IOEither[error, struct{}] {
			return Execute(ctx, spec)
		}),
	)

	return F.Pipe1(
		program(),
		E.Fold(
			F.Identity[error],
			func(struct{}) error { return nil },
		),
	)
}
