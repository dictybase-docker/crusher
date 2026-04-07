package containercreate

import (
	"context"

	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	FP "github.com/cybersiddhu/crush-sandbox/internal/fp"
	"github.com/urfave/cli/v3"
)

// InputFromCommand reads CLI flags and constructs the create Input.
func InputFromCommand(ctx context.Context, cmd *cli.Command) Input {
	return Input{
		Ctx:           ctx,
		ImageName:     cmd.String("image"),
		ContainerName: cmd.String("name"),
		ConfigPath:    cmd.String("config"),
		DataPath:      cmd.String("data"),
		WorkspacePath: cmd.String("workspace"),
		Volumes:       cmd.StringSlice("volume"),
	}
}

// Command returns the CLI definition for the create subcommand.
func Command() *cli.Command {
	return &cli.Command{
		Name:  "create",
		Usage: "Create a Crush sandbox container",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "name",
				Aliases: []string{"n"},
				Usage:   "Container name (auto-generated if omitted)",
			},
			&cli.StringFlag{
				Name:    "image",
				Aliases: []string{"i"},
				Usage:   "Image name",
				Value:   DefaultImageName,
			},
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Usage:   "Host path to Crush config directory (required)",
			},
			&cli.StringFlag{
				Name:    "data",
				Aliases: []string{"d"},
				Usage:   "Host path to Crush data directory (required)",
			},
			&cli.StringFlag{
				Name:    "workspace",
				Aliases: []string{"w"},
				Usage:   "Host path to workspace directory (optional)",
			},
			&cli.StringSliceFlag{
				Name:    "volume",
				Aliases: []string{"v"},
				Usage:   "Additional host path to mount (read-only, repeatable)",
			},
		},
		Action: Action,
	}
}

// Action is the create subcommand entry point.
// Pipeline: validate input → execute container create → return result.
func Action(ctx context.Context, cmd *cli.Command) error {
	return F.Pipe5(
		InputFromCommand(ctx, cmd),
		ValidateInput,
		IOE.FromEither[error],
		IOE.Chain(Execute),
		FP.ToEither[error, ContainerResult],
		E.Fold(
			F.Identity[error],
			func(ContainerResult) error { return nil },
		),
	)
}
