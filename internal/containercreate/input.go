// Package containercreate implements the "create" subcommand, which creates a
// container for running the Crush AI assistant with pre-configured mounts.
package containercreate

import (
	"context"
)

const (
	// ContainerHome is the home directory inside the container.
	ContainerHome = "/home/agent"

	// ConfigTarget is the config mount target inside the container.
	ConfigTarget = ContainerHome + "/crush/config"

	// DataTarget is the data mount target inside the container.
	DataTarget = ContainerHome + "/crush/data"

	// WorkspaceTarget is the workspace mount target inside the container.
	WorkspaceTarget = ContainerHome + "/workspace"

	// DefaultImageName is the default image name matching the build subcommand.
	DefaultImageName = "crusher:latest"
)

// MountSpec represents a single volume mount specification.
type MountSpec struct {
	HostPath   string // Absolute path on host
	TargetPath string // Absolute path inside container
	Readonly   bool   // Whether mount is read-only
}

// Input holds all parameters for container creation.
type Input struct {
	ImageName     string   // Image name (e.g., "crusher:latest")
	ContainerName string   // Container name (empty = auto-generate)
	ConfigPath    string   // Host path to Crush config directory (required)
	DataPath      string   // Host path to Crush data directory (required)
	WorkspacePath string   // Host path to workspace (optional)
	Volumes       []string // Additional host paths to mount (read-only)
	Ctx           context.Context
}

// ResolvedInput is the validated and resolved input with absolutized paths.
type ResolvedInput struct {
	ImageName     string
	ContainerName string
	Mounts        []MountSpec
	Workdir       string // Working directory inside container (empty if no workspace)
}

// ContainerResult holds the result of a successful container creation.
type ContainerResult struct {
	Name string // The container name
}

// CommandSpec holds the resolved executable binary and argv slice.
type CommandSpec struct {
	Bin  string   // "container"
	Args []string // Full argument list for "container create"
}
