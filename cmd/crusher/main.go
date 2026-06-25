package main

import (
	"context"
	"fmt"
	"os"

	"github.com/dictybase-docker/crusher/internal/containerbuild"
	"github.com/dictybase-docker/crusher/internal/containercreate"
	"github.com/dictybase-docker/crusher/internal/containeropencodebx"
	"github.com/dictybase-docker/crusher/internal/containersbx"
	"github.com/urfave/cli/v3"
)

func main() {
	app := &cli.Command{
		Name:  "crusher",
		Usage: "Build OCI images through the docker CLI",
		Commands: []*cli.Command{
			containerbuild.Command(),
			containercreate.Command(),
			containersbx.Command(),
			containeropencodebx.Command(),
		},
	}
	if err := app.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
