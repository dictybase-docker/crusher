package predarrays

import (
	"testing"

	"github.com/stretchr/testify/assert"

	predord "github.com/dictybase-docker/crusher/internal/fputils/predicate/ord"
)

const (
	fpGoHTTP     = "fp-go-http"
	fpGoLens     = "fp-go-lens"
	fpGoPRReview = "fp-go-pr-review"
)

func TestMemberOf_string(t *testing.T) {
	allowedTags := []string{"latest", "stable", "nightly"}
	isAllowedTag := MemberOf(predord.StringEq)

	assert.True(t, isAllowedTag("stable")(allowedTags))
	assert.True(t, isAllowedTag("latest")(allowedTags))
	assert.False(t, isAllowedTag("beta")(allowedTags))
	assert.False(t, isAllowedTag("")(allowedTags))
	assert.False(t, isAllowedTag("latest")(nil))
	assert.False(t, isAllowedTag("latest")([]string{}))
}

func TestNotMemberOf_string(t *testing.T) {
	allowedTags := []string{"latest", "stable", "nightly"}
	isNotAllowedTag := NotMemberOf(predord.StringEq)

	assert.False(t, isNotAllowedTag("stable")(allowedTags))
	assert.False(t, isNotAllowedTag("latest")(allowedTags))
	assert.True(t, isNotAllowedTag("beta")(allowedTags))
	assert.True(t, isNotAllowedTag("")(allowedTags))
	assert.True(t, isNotAllowedTag("latest")(nil))
	assert.True(t, isNotAllowedTag("latest")([]string{}))
}

func TestIntersect_string(t *testing.T) {
	knownSkills := []string{"fp-go", fpGoHTTP, fpGoLens}
	requested := []string{fpGoHTTP, fpGoLens, fpGoPRReview}

	got := Intersect(predord.StringEq)(knownSkills)(requested)

	assert.Equal(t, []string{fpGoHTTP, fpGoLens}, got)
	assert.Empty(t, Intersect(predord.StringEq)(nil)(requested))
	assert.Empty(t, Intersect(predord.StringEq)(knownSkills)(nil))
}

func TestDifference_string(t *testing.T) {
	knownSkills := []string{"fp-go", fpGoHTTP, fpGoLens}
	requested := []string{fpGoHTTP, fpGoLens, fpGoPRReview}

	got := Difference(predord.StringEq)(knownSkills)(requested)

	assert.Equal(t, []string{fpGoPRReview}, got)
	assert.Equal(t, requested, Difference(predord.StringEq)(nil)(requested))
	assert.Empty(t, Difference(predord.StringEq)(knownSkills)(nil))
}
