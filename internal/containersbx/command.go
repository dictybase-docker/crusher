package containersbx

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
	nord4     = color.RGB(0xD8, 0xDE, 0xE9) //nolint:mnd
	nord8     = color.RGB(0x88, 0xC0, 0xD0) //nolint:mnd
	nord10    = color.RGB(0x5E, 0x81, 0xAC) //nolint:mnd
	nord8bold = color.NewPrinter(
		Str.IntersperseSemigroup(";").Concat(nord8.Code(), color.OpBold.Code()),
	)
)

// InputFromCommand reads CLI flags and constructs the sbx Input.
func InputFromCommand(ctx context.Context, cmd *cli.Command) Input {
	return Input{
		Ctx:                 ctx,
		OutputPath:          cmd.String("output"),
		ConfigPath:          cmd.String("config"),
		SkillsPath:          cmd.String("skills"),
		KitName:             cmd.String("name"),
		APIKey:              cmd.String("api-key"),
		ShouldCreate:        cmd.Bool("create"),
		AgentImage:          cmd.String("image"),
		CrushVersion:        cmd.String("crush-version"),
		GolangciLintVersion: cmd.String("golangci-lint-version"),
		GoVersion:           cmd.String("go-version"),
		GotestsumVersion:    cmd.String("gotestsum-version"),
		MoxideVersion:       cmd.String("moxide-version"),
		SemVersion:          cmd.String("sem-version"),
		RtkVersion:          cmd.String("rtk-version"),
	}
}

// Command returns the CLI definition for the sbx subcommand.
func Command() *cli.Command {
	return &cli.Command{
		Name:  "sbx",
		Usage: "Generate, validate, and pack a Docker Sandbox agent kit",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Usage:   "Path for the packed kit zip file",
				Value:   DefaultOutputPath,
			},
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Usage:   "Path to crush.json (default OpenRouter config if omitted)",
			},
			&cli.StringFlag{
				Name:    "skills",
				Aliases: []string{"s"},
				Usage:   "Path to skills directory",
			},
			&cli.StringFlag{
				Name:    "name",
				Aliases: []string{"n"},
				Usage:   "Sandbox display name (default: current directory basename)",
			},
			&cli.StringFlag{
				Name:     "api-key",
				Aliases:  []string{"k"},
				Usage:    "OpenRouter API key",
				Required: true,
			},
			&cli.BoolFlag{
				Name:  createCmd,
				Usage: "Create the sandbox instance after packing",
			},
			&cli.StringFlag{
				Name:  "image",
				Usage: "Base Docker image for the sandbox agent",
			},
			&cli.StringFlag{
				Name:  "crush-version",
				Usage: "Crush version for go install",
				Value: DefaultCrushVersion,
			},
			&cli.StringFlag{
				Name:  "golangci-lint-version",
				Usage: "golangci-lint version",
				Value: DefaultGolangciLintVersion,
			},
			&cli.StringFlag{
				Name:  "go-version",
				Usage: "Go toolchain version",
				Value: DefaultGoVersion,
			},
			&cli.StringFlag{
				Name:  "gotestsum-version",
				Usage: "gotestsum version",
				Value: DefaultGotestsumVersion,
			},
			&cli.StringFlag{
				Name:  "moxide-version",
				Usage: "markdown-oxide version",
				Value: DefaultMoxideVersion,
			},
			&cli.StringFlag{
				Name:  "sem-version",
				Usage: "sem version",
				Value: DefaultSemVersion,
			},
			&cli.StringFlag{
				Name:  "rtk-version",
				Usage: "rtk version",
				Value: DefaultRtkVersion,
			},
		},
		Action: Action,
	}
}

// Action is the sbx subcommand entry point.
// Pipeline: normalize input → validate input → execute kit pipeline → fold into Pair.
// Printing is outside the pipeline.
func Action(ctx context.Context, cmd *cli.Command) error {
	result := F.Pipe5(
		InputFromCommand(ctx, cmd),
		ValidateInput,
		IOE.FromEither[error],
		IOE.Chain(Execute),
		FP.ToEither[error, KitResult],
		E.Fold(
			func(err error) P.Pair[error, KitResult] {
				return P.MakePair(err, KitResult{})
			},
			func(r KitResult) P.Pair[error, KitResult] {
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

func printResult(r KitResult) F.Void {
	nord8bold.Println("✓ Kit validated")
	nord4.Print("✓ Kit packed: ")
	nord8bold.Println(r.OutputPath)

	if r.Created {
		nord8bold.Println("✓ Secret stored for openrouter")
		color.Println()
		nord4.Print("✓ Sandbox created: ")
		nord8bold.Println(r.KitName)
		color.Println()
		nord8.Print("To start: ")
		nord10.Printf("sbx start %s", r.KitName)
	} else {
		color.Println()
		nord8.Print("To run:         ")
		nord10.Printf("sbx run --kit %s crush", r.OutputPath)
	}

	color.Println()

	return F.VOID
}
