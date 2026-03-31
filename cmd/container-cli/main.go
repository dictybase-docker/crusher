package main

import (
	"context"
	"fmt"
	"os"

	"github.com/cybersiddhu/crush-sandbox/internal/build"
	"github.com/urfave/cli/v3"
)

func main() {
	app := &cli.Command{
		Name:  "container-cli",
		Usage: "Build OCI images through the container CLI",
		Commands: []*cli.Command{
			build.Command(),
		},
	}
	if err := app.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
