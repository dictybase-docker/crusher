package build

import _ "embed"

//go:embed Dockerfile
var embeddedDockerfile string
