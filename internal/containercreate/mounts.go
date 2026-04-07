package containercreate

import (
	"fmt"

	A "github.com/IBM/fp-go/v2/array"
	F "github.com/IBM/fp-go/v2/function"
	O "github.com/IBM/fp-go/v2/option"
)

// renderMount converts a MountSpec to a container --mount argument.
// Format: type=bind,source=<host>,target=<container>[,readonly]
func renderMount(mount MountSpec) string {
	base := fmt.Sprintf(
		"type=bind,source=%s,target=%s",
		mount.HostPath,
		mount.TargetPath,
	)
	return F.Pipe2(
		O.FromPredicate(func(bool) bool { return mount.Readonly })(mount.Readonly),
		O.Fold(
			func() string { return base },
			func(bool) string { return base + ",readonly" },
		),
		F.Identity[string],
	)
}

// renderMountArgs converts a MountSpec to a pair of arguments ["--mount", "<spec>"].
func renderMountArgs(mount MountSpec) []string {
	return []string{"--mount", renderMount(mount)}
}

// renderAllMounts converts all MountSpecs to flattened --mount arguments.
func renderAllMounts(mounts []MountSpec) []string {
	return F.Pipe2(
		mounts,
		A.Map(renderMountArgs),
		A.Flatten,
	)
}

// renderEnvVars returns the environment variable arguments for Crush.
func renderEnvVars() []string {
	return []string{
		"--env", fmt.Sprintf("CRUSH_GLOBAL_CONFIG=%s", ConfigTarget),
		"--env", fmt.Sprintf("CRUSH_GLOBAL_DATA=%s", DataTarget),
	}
}
