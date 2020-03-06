package shell

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/gruntwork-io/gruntwork-cli/errors"
)

// Run the specified shell command with the specified arguments. Connect the command's stdin, stdout, and stderr to
// the currently running app.
func RunShellCommand(options *ShellOptions, command string, args ...string) error {
	if options.SensitiveArgs {
		options.Logger.Infof("Running command: %s (args redacted)", command)
	} else {
		options.Logger.Infof("Running command: %s %s", command, strings.Join(args, " "))
	}

	cmd := exec.Command(command, args...)

	// TODO: consider logging this via options.Logger
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Dir = options.WorkingDir

	return errors.WithStackTrace(cmd.Run())
}

// Run the specified shell command with the specified arguments. Return its stdout and stderr as a string
func RunShellCommandAndGetOutput(options *ShellOptions, command string, args ...string) (string, error) {
	if options.SensitiveArgs {
		options.Logger.Infof("Running command: %s (args redacted)", command)
	} else {
		options.Logger.Infof("Running command: %s %s", command, strings.Join(args, " "))
	}

	cmd := exec.Command(command, args...)

	cmd.Stdin = os.Stdin
	cmd.Dir = options.WorkingDir

	out, err := cmd.CombinedOutput()
	return string(out), errors.WithStackTrace(err)
}

// Run the specified shell command with the specified arguments. Return its stdout and stderr as a string and also
// stream stdout and stderr to the OS stdout/stderr
func RunShellCommandAndGetAndStreamOutput(options *ShellOptions, command string, args ...string) (string, error) {
	if options.SensitiveArgs {
		options.Logger.Infof("Running command: %s (args redacted)", command)
	} else {
		options.Logger.Infof("Running command: %s %s", command, strings.Join(args, " "))
	}

	cmd := exec.Command(command, args...)

	cmd.Dir = options.WorkingDir

	cmd.Stdin = os.Stdin

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", errors.WithStackTrace(err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return "", errors.WithStackTrace(err)
	}

	if err := cmd.Start(); err != nil {
		return "", errors.WithStackTrace(err)
	}

	output, err := readStdoutAndStderr(stdout, stderr, options)
	if err != nil {
		return output, err
	}

	err = cmd.Wait()
	return output, errors.WithStackTrace(err)
}

// This function captures stdout and stderr while still printing it to the stdout and stderr of this Go program
func readStdoutAndStderr(stdout io.ReadCloser, stderr io.ReadCloser, options *ShellOptions) (string, error) {
	allOutput := []string{}

	stdoutScanner := bufio.NewScanner(stdout)
	stderrScanner := bufio.NewScanner(stderr)

	for {
		if stdoutScanner.Scan() {
			text := stdoutScanner.Text()
			options.Logger.Println(text)
			allOutput = append(allOutput, text)
		} else if stderrScanner.Scan() {
			text := stderrScanner.Text()
			options.Logger.Println(text)
			allOutput = append(allOutput, text)
		} else {
			break
		}
	}

	if err := stdoutScanner.Err(); err != nil {
		return "", errors.WithStackTrace(err)
	}

	if err := stderrScanner.Err(); err != nil {
		return "", errors.WithStackTrace(err)
	}

	return strings.Join(allOutput, "\n"), nil
}

// Return true if the OS has the given command installed
func CommandInstalled(command string) bool {
	_, err := exec.LookPath(command)
	return err == nil
}

// CommandInstalledE returns an error if command is not installed
func CommandInstalledE(command string) error {
	if commandExists := CommandInstalled(command); !commandExists {
		err := fmt.Errorf("Command %s is not installed", command)
		return errors.WithStackTrace(err)
	}
	return nil
}
