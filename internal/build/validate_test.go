package build

import (
	"errors"
	"testing"

	E "github.com/IBM/fp-go/v2/either"
)

func TestValidateInput_EmptyFile(t *testing.T) {
	req := Input{
		File: "",
		Tags: []string{"latest"},
	}

	result := ValidateInput(req)

	if E.IsRight(result) {
		t.Error("expected Left for empty Dockerfile path")
	}

	var err error
	_ = E.Fold(func(e error) error {
		err = e
		return nil
	}, func(_ Input) error {
		return errors.New("unexpected Right")
	})(result)

	if err == nil {
		t.Error("expected error for empty Dockerfile path")
	}
	if err.Error() != "dockerfile path is required" {
		t.Errorf("unexpected error message: %s", err.Error())
	}
}

func TestValidateInput_EmptyTagList(t *testing.T) {
	req := Input{
		File: "Dockerfile",
		Tags: []string{},
	}

	result := ValidateInput(req)

	if E.IsRight(result) {
		t.Error("expected Left for empty tag list")
	}

	var err error
	_ = E.Fold(func(e error) error {
		err = e
		return nil
	}, func(_ Input) error {
		return errors.New("unexpected Right")
	})(result)

	if err == nil {
		t.Error("expected error for empty tag list")
	}
	if err.Error() != "tag values must be non-empty" {
		t.Errorf("unexpected error message: %s", err.Error())
	}
}

func TestValidateInput_BlankTagEntry(t *testing.T) {
	req := Input{
		File: "Dockerfile",
		Tags: []string{"latest", ""},
	}

	result := ValidateInput(req)

	if E.IsRight(result) {
		t.Error("expected Left for blank tag entry")
	}

	var err error
	_ = E.Fold(func(e error) error {
		err = e
		return nil
	}, func(_ Input) error {
		return errors.New("unexpected Right")
	})(result)

	if err == nil {
		t.Error("expected error for blank tag entry")
	}
	if err.Error() != "tag values must be non-empty" {
		t.Errorf("unexpected error message: %s", err.Error())
	}
}

func TestValidateInput_DefaultInput(t *testing.T) {
	req := Input{
		File: "Dockerfile",
		Tags: []string{"latest"},
	}

	result := ValidateInput(req)

	if E.IsLeft(result) {
		t.Error("expected Right for default input")
	}

	var validated Input
	_ = E.Fold(func(_ error) error {
		return errors.New("unexpected Left")
	}, func(r Input) error {
		validated = r
		return nil
	})(result)

	if validated.File != "Dockerfile" {
		t.Errorf("expected File to be Dockerfile, got %s", validated.File)
	}
	if len(validated.Tags) != 1 || validated.Tags[0] != "latest" {
		t.Errorf("expected Tags to be [latest], got %v", validated.Tags)
	}
}

func TestValidateInput_MultipleTags(t *testing.T) {
	req := Input{
		File: "docker/Prod.Dockerfile",
		Tags: []string{"latest", "stable", "v1.0.0"},
	}

	result := ValidateInput(req)

	if E.IsLeft(result) {
		t.Error("expected Right for valid build input with multiple tags")
	}

	var validated Input
	_ = E.Fold(func(_ error) error {
		return errors.New("unexpected Left")
	}, func(r Input) error {
		validated = r
		return nil
	})(result)

	if validated.File != "docker/Prod.Dockerfile" {
		t.Errorf("expected File to be docker/Prod.Dockerfile, got %s", validated.File)
	}
	if len(validated.Tags) != 3 {
		t.Errorf("expected 3 tags, got %d", len(validated.Tags))
	}
}
