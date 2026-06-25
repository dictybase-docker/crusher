package containeropencodebx

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

const (
	testKitName    = "my-kit"
	testKitName2   = "test-kit"
	testOutputPath = "/tmp/out.zip"
	apiKeyFlag     = "--api-key"
	testAPIKey     = "my-api-key"
	testAPIKey2    = "sk-abc123"
	testGLVersion  = "1.60.0"
)

func TestInputFromCommand_RequiredAPIKey(t *testing.T) {
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

func TestInputFromCommand_DefaultProvider(t *testing.T) {
	require := require.New(t)
	app := &cli.Command{
		Flags: Command().Flags,
		Action: func(ctx context.Context, cmd *cli.Command) error {
			input := InputFromCommand(ctx, cmd)
			require.Equal(testAPIKey, input.APIKey)
			// The --provider flag has a default value, so it is populated
			// directly by InputFromCommand.
			require.Equal(DefaultProvider, input.Provider)
			require.Equal(DefaultOutputPath, input.OutputPath)
			require.Equal(DefaultAgentImage, input.AgentImage)
			require.False(input.ShouldCreate)

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
			require.Equal(testKitName, input.KitName)
			require.Equal(testAPIKey2, input.APIKey)
			require.Equal(providerAnthropic, input.Provider)
			require.True(input.ShouldCreate)
			require.Equal("custom/image:v3", input.AgentImage)
			require.Equal(testGLVersion, input.GolangciLintVersion)

			return nil
		},
	}
	_ = app.Run(context.Background(), []string{
		sbxBinary,
		"--output", "/tmp/output.zip",
		"--name", testKitName,
		apiKeyFlag, testAPIKey2,
		"--provider", providerAnthropic,
		"--create",
		"--image", "custom/image:v3",
		"--golangci-lint-version", testGLVersion,
	})
}

func TestInputFromCommand_ShortFlags(t *testing.T) {
	require := require.New(t)
	app := &cli.Command{
		Flags: Command().Flags,
		Action: func(ctx context.Context, cmd *cli.Command) error {
			input := InputFromCommand(ctx, cmd)
			require.Equal(testOutputPath, input.OutputPath)
			require.Equal(testKitName2, input.KitName)
			require.Equal("sk-key", input.APIKey)
			require.Equal(providerOpenAI, input.Provider)

			return nil
		},
	}
	_ = app.Run(context.Background(), []string{
		sbxBinary,
		"-o", testOutputPath,
		"-n", testKitName2,
		"-k", "sk-key",
		"-p", providerOpenAI,
	})
}

func TestInputFromCommand_ProviderFlag(t *testing.T) {
	require := require.New(t)
	app := &cli.Command{
		Flags: Command().Flags,
		Action: func(ctx context.Context, cmd *cli.Command) error {
			input := InputFromCommand(ctx, cmd)
			require.Equal(providerAnthropic, input.Provider)

			return nil
		},
	}
	_ = app.Run(
		context.Background(),
		[]string{sbxBinary, apiKeyFlag, testAPIKey, "--provider", providerAnthropic},
	)
}
