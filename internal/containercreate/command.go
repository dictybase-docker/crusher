package containercreate

import (
	"context"

	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	P "github.com/IBM/fp-go/v2/pair"
	Str "github.com/IBM/fp-go/v2/string"
	FP "github.com/dictybase-docker/crusher/internal/fp"
	"github.com/gookit/color"
	"github.com/urfave/cli/v3"
)

var (
	nord4     = color.RGB(0xD8, 0xDE, 0xE9) //nolint:mnd // Polarity Snow — labels
	nord8     = color.RGB(0x88, 0xC0, 0xD0) //nolint:mnd // Frost Bright — name anchor
	nord10    = color.RGB(0x5E, 0x81, 0xAC) //nolint:mnd // Frost Dark — soothing command hints
	nord14    = color.RGB(0xA3, 0xBE, 0x8C) //nolint:mnd // Aurora Green — success
	nord8bold = color.NewPrinter(
		Str.IntersperseSemigroup(";").Concat(nord8.Code(), color.OpBold.Code()),
	)
)

// InputFromCommand reads CLI flags and constructs the create Input.
func InputFromCommand(ctx context.Context, cmd *cli.Command) Input {
	return Input{
		Ctx:           ctx,
		ImageName:     cmd.String("image"),
		ContainerName: cmd.String("name"),
		ConfigPath:    cmd.String("config"),
		DataPath:      cmd.String("data"),
		SkillsPath:    cmd.String("skills"),
		APIKey:        cmd.String("api-key"),
		GitHubToken:   cmd.String("github-token"),
		WorkspacePath: cmd.String("workspace"),
		Volumes:       cmd.StringSlice("volume"),
	}
}

// Command returns the CLI definition for the create subcommand.
func Command() *cli.Command {
	return &cli.Command{
		Name:  createCmd,
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
				Name:     "config",
				Aliases:  []string{"c"},
				Usage:    "Host path to Crush config directory",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "data",
				Aliases:  []string{"d"},
				Usage:    "Host path to Crush data directory ",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "skills",
				Aliases:  []string{"s"},
				Usage:    "Host path to Crush skills directory",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "api-key",
				Aliases:  []string{"k"},
				Usage:    "API key for Crush",
				Required: true,
			},
			&cli.StringFlag{
				Name:    "workspace",
				Aliases: []string{"w"},
				Usage:   "Host path to workspace directory, current directory by default",
			},
			&cli.StringFlag{
				Name:    "github-token",
				Aliases: []string{"g"},
				Usage:   "GitHub personal access token",
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
// Pipeline: normalize input → validate input → execute container create → fold into Pair.
// Printing is outside the pipeline.
func Action(ctx context.Context, cmd *cli.Command) error {
	result := F.Pipe6(
		InputFromCommand(ctx, cmd),
		ValidateInput,
		IOE.FromEither[error],
		IOE.Chain(Execute),
		IOE.Chain(StartContainer),
		FP.ToEither[error, ContainerResult],
		E.Fold(
			func(err error) P.Pair[error, ContainerResult] {
				return P.MakePair(err, ContainerResult{})
			},
			func(r ContainerResult) P.Pair[error, ContainerResult] {
				return P.MakePair[error](nil, r)
			},
		),
	)

	if err := P.Head(result); err != nil {
		return err
	}

	F.Pipe2(result, P.Tail, printResult)

	return nil
}

func printResult(r ContainerResult) F.Void {
	nord4.Print("Container ")
	nord8bold.Printf("%q", r.Name)
	nord4.Print(" ")
	nord14.Println("created and started.")

	nord8.Print("Attach with: ")
	nord10.Printf("docker exec -it %s /bin/sh", r.Name)
	color.Println()

	return F.VOID
}
