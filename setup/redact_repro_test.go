package setup

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRedact_Handle_Attributes(t *testing.T) {
	var buf bytes.Buffer
	handler := newRedactHandler(slog.NewJSONHandler(&buf, nil), []string{"password"})
	logger := slog.New(handler)

	// This is where I suspect the bug is:
	// attributes passed directly to Info should be redacted but might not be.
	logger.Info("login attempt", "user", "admin", "password", "secret123")

	var m map[string]any
	err := json.Unmarshal(buf.Bytes(), &m)
	assert.NoError(t, err)

	t.Logf("Log output: %s", buf.String())
	assert.Equal(t, "***", m["password"], "Password should be redacted in Handle")
}
