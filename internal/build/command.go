// Package build provides a CLI command to build an OCI image using the
// container CLI .
package build

import (
	"context"

	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	FP "github.com/cybersiddhu/crush-sandbox/internal/fp"
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

func Action(ctx context.Context, cmd *cli.Command) error {
	return F.Pipe6(
		Request{
			File: cmd.String("file"),
			Tags: cmd.StringSlice("tag"),
			Ctx:  ctx,
		},
		ValidateRequest,
		IOE.FromEither[error],
		IOE.Map[error](RenderCommand),
		IOE.Chain(Execute),
		FP.ToEither[error, struct{}],
		E.Fold(
			F.Identity[error],
			func(struct{}) error { return nil },
		),
	)
}
