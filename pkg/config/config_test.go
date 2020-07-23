package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test_read strings together subtests because this is the one place
// we'are dealing with side effects, in the form of env vars.
// By ensuring run order via t.Run we can make sure the env var
// manipulation of later tests doesn't interfere with the earlier
// file based tests.
func Test_Read(t *testing.T) {
	t.Run("FromFile", func(t *testing.T) {
		conf, err := Read("./test-resources/config-test-complete.yaml")
		assert.Nil(t, err)
		assert.Equal(t, conf.LogLevel, "debug")
		assert.Equal(t, conf.Language, "ruby")
		assert.Equal(t, conf.Git.Username, "user")
		assert.Equal(t, conf.Git.Token, "pass")
		assert.Equal(t, conf.Webhook.Secret, "asupersecretkey")
		assert.Equal(t, conf.Ruby.Authors, "somepeople")
		assert.Equal(t, conf.Ruby.Email, "dev@dev.com")
		assert.Equal(t, conf.Ruby.GemRepoUser, "user")
		assert.Equal(t, conf.Ruby.GemRepoPass, "pass")
		assert.Equal(t, conf.Ruby.GemRepoHost, "https://somegemrepo.com")
		assert.Equal(t, conf.Ruby.GemName, "proto-demo")
		assert.Equal(t, conf.Ruby.GRPCVersion, "1.19.0")
		assert.Equal(t, conf.Ruby.Homepage, "https://github.com/someorg/somerepo")
		assert.Equal(t, conf.Ruby.Publish, false)
	})
	t.Run("FromFileWithEmptyField", func(t *testing.T) {
		conf, err := Read("./test-resources/config-test-missing.yaml")
		assert.Nil(t, err)
		assert.Equal(t, conf.Webhook.Secret, "")
	})
	t.Run("FromEnv", func(t *testing.T) {
		os.Setenv("PF_LANGUAGE", "ruby")
		os.Setenv("PF_GIT_USERNAME", "envuser")
		os.Setenv("PF_RUBY_PUBLISH", "false")
		conf, err := Read("")
		assert.Nil(t, err)
		assert.Equal(t, conf.Language, "ruby")
		assert.Equal(t, conf.Git.Username, "envuser")
		assert.Equal(t, conf.Ruby.Publish, false)
	})
	t.Run("EnvOverridesFile", func(t *testing.T) {
		os.Setenv("PF_GIT_USERNAME", "envuser")
		conf, err := Read("./test-resources/config-test-complete.yaml")
		assert.Nil(t, err)
		assert.Equal(t, conf.Git.Username, "envuser")
	})
	t.Run("EnvDoesNotOverrideWithZeroVals", func(t *testing.T) {
		os.Setenv("PF_GIT_USERNAME", "envuser")
		conf, err := Read("./test-resources/config-test-complete.yaml")
		assert.Nil(t, err)
		assert.Equal(t, conf.Git.Username, "envuser")
		assert.Equal(t, conf.Git.Token, "pass")
	})
}
