package setup

import (
	"github.com/eldius/initial-config-go/configs"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_setupLogs(t *testing.T) {
	t.Run("given a configuration without output configuration should return an error", func(t *testing.T) {
		assert.NotNil(t, setupLogs("app", configs.LogFormatJSON, configs.LogLevelDEBUG, "", false))
	})

	t.Run("given a configuration with file output configuration should return no error", func(t *testing.T) {
		assert.Nil(t, setupLogs("app", configs.LogFormatJSON, configs.LogLevelDEBUG, "my-log-file.log", false))
	})

	t.Run("given a configuration with stdout output configuration should return no error", func(t *testing.T) {
		assert.Nil(t, setupLogs("app", configs.LogFormatJSON, configs.LogLevelDEBUG, "", true))
	})

	t.Run("given a configuration with stdout and file output configuration should return no error", func(t *testing.T) {
		assert.Nil(t, setupLogs("app", configs.LogFormatJSON, configs.LogLevelDEBUG, "my-log-file-2.log", true))
	})
}
