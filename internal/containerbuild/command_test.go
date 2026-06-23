package containerbuild

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
		Action: func(ctx context.Context, cmd *cli.Command) error {
			input := InputFromCommand(ctx, cmd)
			require.Equal("crusher", input.Name)
			require.Equal([]string{defaultTag}, input.Tags)
			require.NotNil(input.BuildArgs)
			require.Equal("2.11.4", input.BuildArgs["GOLANGCI_LINT_VERSION"])
			require.Equal(defaultTag, input.BuildArgs["CRUSH_VERSION"])
			require.Equal(defaultTag, input.BuildArgs["GOTESTSUM_VERSION"])
			require.Equal(defaultTag, input.BuildArgs["MOXIDE_VERSION"])
			require.Equal(defaultTag, input.BuildArgs["SEM_VERSION"])
			require.Equal(defaultTag, input.BuildArgs["RTK_VERSION"])

			return nil
		},
	}
	_ = app.Run(context.Background(), []string{buildCmd})
}

func TestInputFromCommand_CustomNameAndTags(t *testing.T) {
	require := require.New(t)
	app := &cli.Command{
		Flags: Command().Flags,
		Action: func(ctx context.Context, cmd *cli.Command) error {
			input := InputFromCommand(ctx, cmd)
			require.Equal("myimage", input.Name)
			require.Equal([]string{"v1", "v2"}, input.Tags)

			return nil
		},
	}
	_ = app.Run(
		context.Background(),
		[]string{buildCmd, "--name", "myimage", "--tag", "v1", "--tag", "v2"},
	)
}

func TestInputFromCommand_EmbedFlag(t *testing.T) {
	require := require.New(t)
	app := &cli.Command{
		Flags: Command().Flags,
		Action: func(ctx context.Context, cmd *cli.Command) error {
			require.True(cmd.Bool("embed"))
			input := InputFromCommand(ctx, cmd)
			require.NotNil(input.DockerfileSource)

			return nil
		},
	}
	_ = app.Run(context.Background(), []string{buildCmd, "--embed"})
}

func TestInputFromCommand_CustomFileFlag(t *testing.T) {
	require := require.New(t)
	app := &cli.Command{
		Flags: Command().Flags,
		Action: func(ctx context.Context, cmd *cli.Command) error {
			input := InputFromCommand(ctx, cmd)
			require.NotNil(input.DockerfileSource)

			return nil
		},
	}
	_ = app.Run(context.Background(), []string{buildCmd, "--file", "/custom/Dockerfile"})
}

func TestInputFromCommand_VersionFlags(t *testing.T) {
	require := require.New(t)
	app := &cli.Command{
		Flags: Command().Flags,
		Action: func(ctx context.Context, cmd *cli.Command) error {
			input := InputFromCommand(ctx, cmd)
			require.Equal("3.0.0", input.BuildArgs["GOLANGCI_LINT_VERSION"])
			require.Equal("v2", input.BuildArgs["CRUSH_VERSION"])
			require.Equal("1.5.0", input.BuildArgs["GOTESTSUM_VERSION"])

			return nil
		},
	}
	_ = app.Run(context.Background(), []string{
		buildCmd,
		"--golangci-lint-version", "3.0.0",
		"--crush-version", "v2",
		"--gotestsum-version", "1.5.0",
	})
}
