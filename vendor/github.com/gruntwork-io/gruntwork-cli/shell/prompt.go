package shell

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/bgentry/speakeasy"
	"github.com/fatih/color"

	"github.com/gruntwork-io/gruntwork-cli/errors"
)

var BRIGHT_GREEN = color.New(color.FgHiGreen, color.Bold)

// Prompt the user for text in the CLI. Returns the text entered by the user.
func PromptUserForInput(prompt string, options *ShellOptions) (string, error) {
	return FPromptUserForInput(os.Stdout, os.Stdin, prompt, options)
}

func FPromptUserForInput(out io.Writer, in io.Reader, prompt string, options *ShellOptions) (string, error) {
	BRIGHT_GREEN.Fprint(out, prompt)

	if options.NonInteractive {
		fmt.Fprintln(out)
		options.Logger.Info("The non-interactive flag is set to true, so assuming 'yes' for all prompts")
		return "yes", nil
	}

	reader := bufio.NewReader(in)

	text, err := reader.ReadString('\n')
	if err != nil {
		return "", errors.WithStackTrace(err)
	}

	return strings.TrimSpace(text), nil
}

// Prompt the user for a yes/no response and return true if they entered yes.
func PromptUserForYesNo(prompt string, options *ShellOptions) (bool, error) {
	return FPromptUserForYesNo(os.Stdout, os.Stdin, prompt, options)
}

func FPromptUserForYesNo(out io.Writer, in io.Reader, prompt string, options *ShellOptions) (bool, error) {
	resp, err := FPromptUserForInput(out, in, fmt.Sprintf("%s (y/n) ", prompt), options)

	if err != nil {
		return false, errors.WithStackTrace(err)
	}

	switch strings.ToLower(resp) {
	case "y", "yes":
		return true, nil
	default:
		return false, nil
	}
}

// Prompt a user for a password or other sensitive info that should not be echoed back to stdout.
func PromptUserForPassword(prompt string, options *ShellOptions) (string, error) {
	BRIGHT_GREEN.Print(prompt)

	if options.NonInteractive {
		return "", errors.WithStackTrace(NonInteractivePasswordPrompt)
	}

	password, err := speakeasy.Ask("")
	if err != nil {
		return "", errors.WithStackTrace(err)
	}

	return password, nil
}

// Custom error types

var NonInteractivePasswordPrompt = fmt.Errorf("The non-interactive flag is set, so unable to prompt user for a password.")
