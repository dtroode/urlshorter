package logger_test

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/dtroode/urlshorter/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLog(t *testing.T) {
	log := logger.NewLog("info")
	require.NotNil(t, log)
}

func TestLogger_Fatal(t *testing.T) {
	if os.Getenv("BE_FATAL") == "1" {
		log := logger.NewLog("info")
		log.Fatal("test message")
		return
	}

	var buf bytes.Buffer
	cmd := exec.Command(os.Args[0], "-test.run=TestLogger_Fatal")
	cmd.Env = append(os.Environ(), "BE_FATAL=1")
	cmd.Stderr = &buf
	cmd.Stdout = &buf

	err := cmd.Run()
	e, ok := err.(*exec.ExitError)
	require.True(t, ok, "err should be of type *exec.ExitError")
	require.False(t, e.Success(), "process should exit with error")

	output := buf.String()
	assert.True(t, strings.Contains(output, "test message"), "log output should contain the message")

	var logEntry map[string]interface{}
	err = json.Unmarshal([]byte(output), &logEntry)
	require.NoError(t, err)
	assert.Equal(t, "ERROR", logEntry["level"])
	assert.Equal(t, "test message", logEntry["msg"])
}
