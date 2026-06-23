package containercreate

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

const (
	testSkillsPath = "/host/skills"
	testAPIKeyAlt  = "test-api-key"
)

func TestInputFromCommand_Defaults(t *testing.T) {
	require := require.New(t)
	app := &cli.Command{
		Flags: Command().Flags,
		Action: func(_ context.Context, _ *cli.Command) error {
			require.Fail("should not reach action without required flags")
			return nil
		},
	}
	err := app.Run(context.Background(), []string{createCmd})
	require.Error(err)
	require.Contains(err.Error(), "config")
}

func TestInputFromCommand_MinimalInput(t *testing.T) {
	require := require.New(t)
	app := &cli.Command{
		Flags: Command().Flags,
		Action: func(ctx context.Context, cmd *cli.Command) error {
			input := InputFromCommand(ctx, cmd)
			require.Equal(DefaultImageName, input.ImageName)
			require.Equal(testConfigPath, input.ConfigPath)
			require.Equal(testDataPath, input.DataPath)
			require.Equal(testSkillsPath, input.SkillsPath)
			require.Equal(testAPIKeyAlt, input.APIKey)
			require.Empty(input.ContainerName)

			return nil
		},
	}
	_ = app.Run(context.Background(), []string{
		createCmd,
		configFlag, testConfigPath,
		dataFlag, testDataPath,
		skillsFlag, testSkillsPath,
		apiKeyFlag, testAPIKeyAlt,
	})
}

func TestInputFromCommand_AllFlags(t *testing.T) {
	require := require.New(t)
	app := &cli.Command{
		Flags: Command().Flags,
		Action: func(ctx context.Context, cmd *cli.Command) error {
			input := InputFromCommand(ctx, cmd)
			require.Equal("myimage:v2", input.ImageName)
			require.Equal(testContainerName, input.ContainerName)
			require.Equal(testConfigPath, input.ConfigPath)
			require.Equal(testDataPath, input.DataPath)
			require.Equal(testSkillsPath, input.SkillsPath)
			require.Equal(testAPIKeyAlt, input.APIKey)
			require.Equal("ghp_token", input.GitHubToken)
			require.Equal("/host/workspace", input.WorkspacePath)

			return nil
		},
	}
	_ = app.Run(context.Background(), []string{
		createCmd,
		"--image", "myimage:v2",
		nameFlag, testContainerName,
		configFlag, testConfigPath,
		dataFlag, testDataPath,
		skillsFlag, testSkillsPath,
		apiKeyFlag, testAPIKeyAlt,
		"--github-token", "ghp_token",
		"--workspace", "/host/workspace",
	})
}

func TestInputFromCommand_VolumesFlag(t *testing.T) {
	require := require.New(t)
	app := &cli.Command{
		Flags: Command().Flags,
		Action: func(ctx context.Context, cmd *cli.Command) error {
			input := InputFromCommand(ctx, cmd)
			require.Equal([]string{"/host/vol1", "/host/vol2"}, input.Volumes)

			return nil
		},
	}
	_ = app.Run(context.Background(), []string{
		createCmd,
		configFlag, testConfigPath,
		dataFlag, testDataPath,
		skillsFlag, testSkillsPath,
		apiKeyFlag, testAPIKeyAlt,
		"--volume", "/host/vol1",
		"--volume", "/host/vol2",
	})
}

func TestInputFromCommand_ImageShortFlag(t *testing.T) {
	require := require.New(t)
	app := &cli.Command{
		Flags: Command().Flags,
		Action: func(ctx context.Context, cmd *cli.Command) error {
			input := InputFromCommand(ctx, cmd)
			require.Equal("myimage:latest", input.ImageName)

			return nil
		},
	}
	_ = app.Run(context.Background(), []string{
		createCmd,
		"-i", "myimage:latest",
		configFlag, testConfigPath,
		dataFlag, testDataPath,
		skillsFlag, testSkillsPath,
		apiKeyFlag, testAPIKeyAlt,
	})
}
