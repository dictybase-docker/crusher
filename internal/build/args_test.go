package build

import (
	"strings"
	"testing"
)

func TestRenderCommand_DefaultRequest(t *testing.T) {
	req := Request{
		File: "Dockerfile",
		Tags: []string{"latest"},
	}

	spec := RenderCommand(req)

	if spec.Name != "container" {
		t.Errorf("expected Name to be container, got %s", spec.Name)
	}

	expected := "container build --file Dockerfile --tag latest ."
	actual := spec.Name + " " + strings.Join(spec.Args, " ")
	if actual != expected {
		t.Errorf("expected %q, got %q", expected, actual)
	}
}

func TestRenderCommand_RepeatedTags(t *testing.T) {
	req := Request{
		File: "Dockerfile",
		Tags: []string{"latest", "stable", "v1.0.0"},
	}

	spec := RenderCommand(req)

	if spec.Name != "container" {
		t.Errorf("expected Name to be container, got %s", spec.Name)
	}

	expected := "container build --file Dockerfile --tag latest --tag stable --tag v1.0.0 ."
	actual := spec.Name + " " + strings.Join(spec.Args, " ")
	if actual != expected {
		t.Errorf("expected %q, got %q", expected, actual)
	}
}

func TestRenderCommand_DockerfileOverride(t *testing.T) {
	req := Request{
		File: "docker/Prod.Dockerfile",
		Tags: []string{"latest"},
	}

	spec := RenderCommand(req)

	if spec.Name != "container" {
		t.Errorf("expected Name to be container, got %s", spec.Name)
	}

	expected := "container build --file docker/Prod.Dockerfile --tag latest ."
	actual := spec.Name + " " + strings.Join(spec.Args, " ")
	if actual != expected {
		t.Errorf("expected %q, got %q", expected, actual)
	}
}

func TestRenderCommand_FinalArgIsBuildContext(t *testing.T) {
	req := Request{
		File: "Dockerfile",
		Tags: []string{"latest"},
	}

	spec := RenderCommand(req)

	if len(spec.Args) == 0 {
		t.Error("expected non-empty Args")
		return
	}

	lastArg := spec.Args[len(spec.Args)-1]
	if lastArg != "." {
		t.Errorf("expected final arg to be ., got %q", lastArg)
	}
}

func TestRenderCommand_ArgsOrder(t *testing.T) {
	req := Request{
		File: "Dockerfile",
		Tags: []string{"latest", "stable"},
	}

	spec := RenderCommand(req)

	if len(spec.Args) < 7 {
		t.Errorf("expected at least 7 args, got %d", len(spec.Args))
		return
	}

	if spec.Args[0] != "build" {
		t.Errorf("expected first arg to be build, got %q", spec.Args[0])
	}
	if spec.Args[1] != "--file" {
		t.Errorf("expected second arg to be --file, got %q", spec.Args[1])
	}
	if spec.Args[2] != "Dockerfile" {
		t.Errorf("expected third arg to be Dockerfile, got %q", spec.Args[2])
	}
	if spec.Args[3] != "--tag" {
		t.Errorf("expected fourth arg to be --tag, got %q", spec.Args[3])
	}
	if spec.Args[4] != "latest" {
		t.Errorf("expected fifth arg to be latest, got %q", spec.Args[4])
	}
	if spec.Args[5] != "--tag" {
		t.Errorf("expected sixth arg to be --tag, got %q", spec.Args[5])
	}
	if spec.Args[6] != "stable" {
		t.Errorf("expected seventh arg to be stable, got %q", spec.Args[6])
	}
}
