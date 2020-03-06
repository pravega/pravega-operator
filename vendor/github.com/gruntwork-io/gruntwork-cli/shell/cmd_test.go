package shell

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/gruntwork-io/gruntwork-cli/logging"
)

func TestRunShellCommand(t *testing.T) {
	t.Parallel()

	assert.Nil(t, RunShellCommand(NewShellOptions(), "echo", "hi"))
}

func TestRunShellCommandInvalidCommand(t *testing.T) {
	t.Parallel()

	assert.NotNil(t, RunShellCommand(NewShellOptions(), "not-a-real-command"))
}

func TestRunShellCommandAndGetOutput(t *testing.T) {
	t.Parallel()

	out, err := RunShellCommandAndGetOutput(NewShellOptions(), "echo", "hi")
	assert.Nil(t, err, "Unexpected error: %v", err)
	assert.Equal(t, "hi\n", out)
}

func TestCommandInstalledOnValidCommand(t *testing.T) {
	t.Parallel()

	assert.True(t, CommandInstalled("echo"))
}

func TestCommandInstalledOnInvalidCommand(t *testing.T) {
	t.Parallel()

	assert.False(t, CommandInstalled("not-a-real-command"))
}

func TestCommandInstalledEOnValidCommand(t *testing.T) {
	t.Parallel()

	assert.Nil(t, CommandInstalledE("echo"))
}

func TestCommandInstalledEOnInvalidCommand(t *testing.T) {
	t.Parallel()

	assert.NotNil(t, CommandInstalledE("not-a-real-command"))
}

// Test that when SensitiveArgs is true, do not log the args
func TestSensitiveArgsTrueHidesOnRunShellCommand(t *testing.T) {
	t.Parallel()

	buffer := bytes.NewBufferString("")
	logger := logging.GetLogger("")
	logger.Out = buffer
	options := NewShellOptions()
	options.SensitiveArgs = true
	options.Logger = logger

	assert.Nil(t, RunShellCommand(options, "echo", "hi"))
	assert.NotContains(t, buffer.String(), "hi")
	assert.Contains(t, buffer.String(), "echo")
}

// Test that when SensitiveArgs is false, log the args
func TestSensitiveArgsFalseShowsOnRunShellCommand(t *testing.T) {
	t.Parallel()

	buffer := bytes.NewBufferString("")
	logger := logging.GetLogger("")
	logger.Out = buffer
	options := NewShellOptions()
	options.Logger = logger

	assert.Nil(t, RunShellCommand(options, "echo", "hi"))
	assert.Contains(t, buffer.String(), "hi")
	assert.Contains(t, buffer.String(), "echo")
}

// Test that when SensitiveArgs is true, do not log the args
func TestSensitiveArgsTrueHidesOnRunShellCommandAndGetOutput(t *testing.T) {
	t.Parallel()

	buffer := bytes.NewBufferString("")
	logger := logging.GetLogger("")
	logger.Out = buffer
	options := NewShellOptions()
	options.SensitiveArgs = true
	options.Logger = logger

	_, err := RunShellCommandAndGetOutput(options, "echo", "hi")
	assert.Nil(t, err)
	assert.NotContains(t, buffer.String(), "hi")
	assert.Contains(t, buffer.String(), "echo")
}

// Test that when SensitiveArgs is false, log the args
func TestSensitiveArgsFalseShowsOnRunShellCommandAndGetOutput(t *testing.T) {
	t.Parallel()

	buffer := bytes.NewBufferString("")
	logger := logging.GetLogger("")
	logger.Out = buffer
	options := NewShellOptions()
	options.Logger = logger

	_, err := RunShellCommandAndGetOutput(options, "echo", "hi")
	assert.Nil(t, err)
	assert.Contains(t, buffer.String(), "hi")
	assert.Contains(t, buffer.String(), "echo")
}
