package helm

import (
	"testing"

	"github.com/gruntwork-io/gruntwork-cli/errors"
	"github.com/gruntwork-io/terratest/modules/shell"
)

// getCommonArgs extracts common helm options. In this case, these are:
// - kubeconfig path
// - kubeconfig context
// - helm home path
func getCommonArgs(options *Options, args ...string) []string {
	if options.KubectlOptions != nil && options.KubectlOptions.ContextName != "" {
		args = append(args, "--kube-context", options.KubectlOptions.ContextName)
	}
	if options.KubectlOptions != nil && options.KubectlOptions.ConfigPath != "" {
		args = append(args, "--kubeconfig", options.KubectlOptions.ConfigPath)
	}
	if options.HomePath != "" {
		args = append(args, "--home", options.HomePath)
	}
	return args
}

// getValuesArgsE computes the args to pass in for setting values
func getValuesArgsE(t *testing.T, options *Options, args ...string) ([]string, error) {
	args = append(args, formatSetValuesAsArgs(options.SetValues)...)

	valuesFilesArgs, err := formatValuesFilesAsArgsE(t, options.ValuesFiles)
	if err != nil {
		return args, errors.WithStackTrace(err)
	}
	args = append(args, valuesFilesArgs...)

	setFilesArgs, err := formatSetFilesAsArgsE(t, options.SetFiles)
	if err != nil {
		return args, errors.WithStackTrace(err)
	}
	args = append(args, setFilesArgs...)
	return args, nil
}

// RunHelmCommandAndGetOutputE runs helm with the given arguments and options and returns stdout/stderr.
func RunHelmCommandAndGetOutputE(t *testing.T, options *Options, cmd string, additionalArgs ...string) (string, error) {
	args := []string{cmd}
	args = getCommonArgs(options, args...)
	args = append(args, additionalArgs...)

	helmCmd := shell.Command{
		Command:    "helm",
		Args:       args,
		WorkingDir: ".",
		Env:        options.EnvVars,
	}
	return shell.RunCommandAndGetOutputE(t, helmCmd)
}
