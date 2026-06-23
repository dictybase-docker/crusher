package containersbx

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

const (
	testKitName       = "my-kit"
	testKitName2      = "test-kit"
	testOutputPath    = "/tmp/out.zip"
	tempDirPath       = "/tmp/sbx-test"
	kittempZipPath    = "/tmp/kit.zip"
	testGLVersion     = "2.0.0"
	testCrushVersion2 = "v2.0.0"
	testCrushVersion3 = "v3.0.0"
	apiKeyFlag        = "--api-key"
	testAPIKey        = "my-api-key"
	testAPIKey2       = "sk-abc123"
	testAPIKeyDefault = "test-key"
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
	err := app.Run(context.Background(), []string{sbxBinary})
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
			require.Equal(testAPIKey, input.APIKey)
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
	_ = app.Run(context.Background(), []string{sbxBinary, apiKeyFlag, testAPIKey})
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
			require.Equal(testKitName, input.KitName)
			require.Equal(testAPIKey2, input.APIKey)
			require.True(input.ShouldCreate)
			require.Equal(testCrushVersion3, input.CrushVersion)
			require.Equal("custom/image:v3", input.AgentImage)
			require.Equal("1.60.0", input.GolangciLintVersion)
			require.Equal("1.24.0", input.GoVersion)
			require.Equal("1.8.0", input.GotestsumVersion)

			return nil
		},
	}
	_ = app.Run(context.Background(), []string{
		sbxBinary,
		"--output", "/tmp/output.zip",
		"--config", "/path/to/crush.json",
		"--skills", "/path/to/skills",
		"--name", testKitName,
		apiKeyFlag, testAPIKey2,
		"--create",
		"--image", "custom/image:v3",
		"--crush-version", testCrushVersion3,
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
	_ = app.Run(context.Background(), []string{sbxBinary, apiKeyFlag, testAPIKey})
}

func TestInputFromCommand_ShortFlags(t *testing.T) {
	require := require.New(t)
	app := &cli.Command{
		Flags: Command().Flags,
		Action: func(ctx context.Context, cmd *cli.Command) error {
			input := InputFromCommand(ctx, cmd)
			require.Equal(testOutputPath, input.OutputPath)
			require.Equal("/path/cfg.json", input.ConfigPath)
			require.Equal("/path/skills", input.SkillsPath)
			require.Equal(testKitName2, input.KitName)
			require.Equal("sk-key", input.APIKey)

			return nil
		},
	}
	_ = app.Run(context.Background(), []string{
		sbxBinary,
		"-o", testOutputPath,
		"-c", "/path/cfg.json",
		"-s", "/path/skills",
		"-n", testKitName2,
		"-k", "sk-key",
	})
}
