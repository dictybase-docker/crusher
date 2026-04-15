package containercreate

import (
	"fmt"

	A "github.com/IBM/fp-go/v2/array"
	F "github.com/IBM/fp-go/v2/function"
	O "github.com/IBM/fp-go/v2/option"
)

// renderMount converts a MountSpec to a container --mount argument.
// Format: type=bind,source=<host>,target=<container>[,readonly]
func renderMount(mount MountSpec) []string {
	base := fmt.Sprintf(
		"type=bind,source=%s,target=%s",
		mount.HostPath,
		mount.TargetPath,
	)
	return F.Pipe3(
		mount.Readonly,
		O.FromPredicate(func(bool) bool { return mount.Readonly }),
		O.Fold(
			func() string { return base },
			func(bool) string {
				return mountJoin.Concat(base, "readonly")
			},
		),
		func(spec string) []string {
			return A.From("--mount", spec)
		},
	)
}

// renderEnvVars returns the environment variable arguments for Crush.
func renderEnvVars(apiKey string, githubToken string) []string {
	base := []string{
		"--env", fmt.Sprintf("CRUSH_GLOBAL_CONFIG=%s", ConfigTarget),
		"--env", fmt.Sprintf("CRUSH_GLOBAL_DATA=%s", DataTarget),
		"--env", fmt.Sprintf("OPENROUTER_API_KEY=%s", apiKey),
	}

	return F.Pipe2(
		githubToken,
		O.FromPredicate(func(s string) bool { return s != "" }),
		O.Fold(
			func() []string { return base },
			func(token string) []string {
				return append(base, "--env", fmt.Sprintf("GITHUB_TOKEN=%s", token))
			},
		),
	)
}
