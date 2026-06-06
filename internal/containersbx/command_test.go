package containersbx

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

func TestInputFromCommand_Defaults(t *testing.T) {
	require := require.New(t)
	app := &cli.Command{
		Flags: Command().Flags,
		Action: func(_ context.Context, _ *cli.Command) error {
			require.Fail("should not reach action without required api-key")
			return nil
		},
	}
	err := app.Run(context.Background(), []string{"sbx"})
	require.Error(err)
	require.Contains(err.Error(), "api-key")
}

func TestInputFromCommand_MinimalInput(t *testing.T) {
	require := require.New(t)
	app := &cli.Command{
		Flags: Command().Flags,
		Action: func(ctx context.Context, cmd *cli.Command) error {
			input := InputFromCommand(ctx, cmd)
			require.Equal(DefaultOutputPath, input.OutputPath)
			require.Equal("my-api-key", input.APIKey)
			require.Empty(input.ConfigPath)
			require.Empty(input.SkillsPath)
			require.Empty(input.KitName)
			require.False(input.ShouldCreate)
			require.Equal(DefaultCrushVersion, input.CrushVersion)
			require.Equal(DefaultGolangciLintVersion, input.GolangciLintVersion)
			require.Equal(DefaultGoVersion, input.GoVersion)
			require.Equal(DefaultGotestsumVersion, input.GotestsumVersion)
			require.Equal(DefaultMoxideVersion, input.MoxideVersion)
			require.Equal(DefaultSemVersion, input.SemVersion)
			require.Equal(DefaultRtkVersion, input.RtkVersion)

			return nil
		},
	}
	_ = app.Run(context.Background(), []string{"sbx", "--api-key", "my-api-key"})
}

func TestInputFromCommand_AllFlags(t *testing.T) {
	require := require.New(t)
	app := &cli.Command{
		Flags: Command().Flags,
		Action: func(ctx context.Context, cmd *cli.Command) error {
			input := InputFromCommand(ctx, cmd)
			require.Equal("/tmp/output.zip", input.OutputPath)
			require.Equal("/path/to/crush.json", input.ConfigPath)
			require.Equal("/path/to/skills", input.SkillsPath)
			require.Equal("my-kit", input.KitName)
			require.Equal("sk-abc123", input.APIKey)
			require.True(input.ShouldCreate)
			require.Equal("v3.0.0", input.CrushVersion)
			require.Equal("custom/image:v3", input.AgentImage)
			require.Equal("1.60.0", input.GolangciLintVersion)
			require.Equal("1.24.0", input.GoVersion)
			require.Equal("1.8.0", input.GotestsumVersion)

			return nil
		},
	}
	_ = app.Run(context.Background(), []string{
		"sbx",
		"--output", "/tmp/output.zip",
		"--config", "/path/to/crush.json",
		"--skills", "/path/to/skills",
		"--name", "my-kit",
		"--api-key", "sk-abc123",
		"--create",
		"--image", "custom/image:v3",
		"--crush-version", "v3.0.0",
		"--golangci-lint-version", "1.60.0",
		"--go-version", "1.24.0",
		"--gotestsum-version", "1.8.0",
	})
}

func TestInputFromCommand_ShouldCreateFalse(t *testing.T) {
	require := require.New(t)
	app := &cli.Command{
		Flags: Command().Flags,
		Action: func(ctx context.Context, cmd *cli.Command) error {
			input := InputFromCommand(ctx, cmd)
			require.False(input.ShouldCreate)

			return nil
		},
	}
	_ = app.Run(context.Background(), []string{"sbx", "--api-key", "my-api-key"})
}

func TestInputFromCommand_ShortFlags(t *testing.T) {
	require := require.New(t)
	app := &cli.Command{
		Flags: Command().Flags,
		Action: func(ctx context.Context, cmd *cli.Command) error {
			input := InputFromCommand(ctx, cmd)
			require.Equal("/tmp/out.zip", input.OutputPath)
			require.Equal("/path/cfg.json", input.ConfigPath)
			require.Equal("/path/skills", input.SkillsPath)
			require.Equal("test-kit", input.KitName)
			require.Equal("sk-key", input.APIKey)

			return nil
		},
	}
	_ = app.Run(context.Background(), []string{
		"sbx",
		"-o", "/tmp/out.zip",
		"-c", "/path/cfg.json",
		"-s", "/path/skills",
		"-n", "test-kit",
		"-k", "sk-key",
	})
}
