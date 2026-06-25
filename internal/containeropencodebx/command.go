package containeropencodebx

import (
	"context"

	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	O "github.com/IBM/fp-go/v2/option"
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

// InputFromCommand reads CLI flags and constructs the opencode-sbx Input.
func InputFromCommand(ctx context.Context, cmd *cli.Command) Input {
	return Input{
		Ctx:                 ctx,
		OutputPath:          cmd.String("output"),
		KitName:             cmd.String("name"),
		APIKey:              cmd.String("api-key"),
		Provider:            cmd.String("provider"),
		ShouldCreate:        cmd.Bool("create"),
		AgentImage:          cmd.String("image"),
		GolangciLintVersion: cmd.String("golangci-lint-version"),
		GotestsumVersion:    cmd.String("gotestsum-version"),
		MoxideVersion:       cmd.String("moxide-version"),
		SemVersion:          cmd.String("sem-version"),
		RtkVersion:          cmd.String("rtk-version"),
	}
}

// Command returns the CLI definition for the opencode-sbx subcommand.
func Command() *cli.Command {
	return &cli.Command{
		Name:  "opencode-sbx",
		Usage: "Generate, validate, and pack a Docker Sandbox agent kit for opencode",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Usage:   "Path for the packed kit zip file",
				Value:   DefaultOutputPath,
			},
			&cli.StringFlag{
				Name:    "name",
				Aliases: []string{"n"},
				Usage:   "Sandbox display name (default: auto-generated)",
			},
			&cli.StringFlag{
				Name:     "api-key",
				Aliases:  []string{"k"},
				Usage:    "AI provider API key",
				Required: true,
			},
			&cli.StringFlag{
				Name:    "provider",
				Aliases: []string{"p"},
				Usage:   "AI provider: openrouter, anthropic, openai, google",
				Value:   DefaultProvider,
			},
			&cli.BoolFlag{
				Name:  "create",
				Usage: "Create the sandbox instance after packing",
			},
			&cli.StringFlag{
				Name:  "image",
				Usage: "Base Docker image for the sandbox agent",
				Value: DefaultAgentImage,
			},
			&cli.StringFlag{
				Name:  "golangci-lint-version",
				Usage: "golangci-lint version",
				Value: DefaultGolangciLintVersion,
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

// Action is the opencode-sbx subcommand entry point.
func Action(ctx context.Context, cmd *cli.Command) error {
	return F.Pipe5(
		InputFromCommand(ctx, cmd),
		ValidateInput,
		IOE.FromEither[error],
		IOE.Chain(Execute),
		FP.ToEither[error, KitResult],
		E.Fold(
			F.Identity[error],
			func(r KitResult) error {
				printResult(r)
				return nil
			},
		),
	)
}

// printResult writes the kit result to the console. Branches on r.Created via
// Option combinators to honour the no-imperative-branching rule.
func printResult(r KitResult) {
	nord8bold.Println("✓ Kit validated")
	nord4.Print("✓ Kit packed: ")
	nord8bold.Println(r.OutputPath)

	F.Pipe2(
		r.Created,
		O.FromPredicate(F.Identity[bool]),
		O.Fold[bool, any](
			func() any {
				color.Println()
				nord8.Print("To run:  ")
				nord10.Printf("sbx run --kit %s opencode", r.OutputPath)

				return nil
			},
			func(_ bool) any {
				nord8bold.Println("✓ Secret stored for " + r.KitName)
				color.Println()
				nord4.Print("✓ Sandbox created: ")
				nord8bold.Println(r.KitName)
				color.Println()
				nord8.Print("To start: ")
				nord10.Printf("sbx start %s", r.KitName)

				return nil
			},
		),
	)
	color.Println()
}
