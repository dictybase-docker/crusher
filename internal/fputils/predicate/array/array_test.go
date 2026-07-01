package predarrays

import (
	"testing"

	"github.com/stretchr/testify/assert"

	predord "github.com/dictybase-docker/crusher/internal/fputils/predicate/ord"
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