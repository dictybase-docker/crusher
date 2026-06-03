package containercreate

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
			require.Fail("should not reach action without required flags")
			return nil
		},
	}
	err := app.Run(context.Background(), []string{"create"})
	require.Error(err)
	require.Contains(err.Error(), "config")
}

func TestInputFromCommand_MinimalInput(t *testing.T) {
	require := require.New(t)
	app := &cli.Command{
		Flags: Command().Flags,
		Action: func(ctx context.Context, cmd *cli.Command) error {
			input := InputFromCommand(ctx, cmd)
			require.Equal("crusher:latest", input.ImageName)
			require.Equal("/host/config", input.ConfigPath)
			require.Equal("/host/data", input.DataPath)
			require.Equal("/host/skills", input.SkillsPath)
			require.Equal("test-api-key", input.APIKey)
			require.Empty(input.ContainerName)

			return nil
		},
	}
	_ = app.Run(context.Background(), []string{
		"create",
		"--config", "/host/config",
		"--data", "/host/data",
		"--skills", "/host/skills",
		"--api-key", "test-api-key",
	})
}

func TestInputFromCommand_AllFlags(t *testing.T) {
	require := require.New(t)
	app := &cli.Command{
		Flags: Command().Flags,
		Action: func(ctx context.Context, cmd *cli.Command) error {
			input := InputFromCommand(ctx, cmd)
			require.Equal("myimage:v2", input.ImageName)
			require.Equal("mycontainer", input.ContainerName)
			require.Equal("/host/config", input.ConfigPath)
			require.Equal("/host/data", input.DataPath)
			require.Equal("/host/skills", input.SkillsPath)
			require.Equal("test-api-key", input.APIKey)
			require.Equal("ghp_token", input.GitHubToken)
			require.Equal("/host/workspace", input.WorkspacePath)

			return nil
		},
	}
	_ = app.Run(context.Background(), []string{
		"create",
		"--image", "myimage:v2",
		"--name", "mycontainer",
		"--config", "/host/config",
		"--data", "/host/data",
		"--skills", "/host/skills",
		"--api-key", "test-api-key",
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
		"create",
		"--config", "/host/config",
		"--data", "/host/data",
		"--skills", "/host/skills",
		"--api-key", "test-api-key",
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
		"create",
		"-i", "myimage:latest",
		"--config", "/host/config",
		"--data", "/host/data",
		"--skills", "/host/skills",
		"--api-key", "test-api-key",
	})
}
