package entrypoint

import (
	"fmt"
	"os"

	"github.com/urfave/cli"

	"github.com/gruntwork-io/gruntwork-cli/errors"
	"github.com/gruntwork-io/gruntwork-cli/logging"
)

const defaultSuccessExitCode = 0
const defaultErrorExitCode = 1
const debugEnvironmentVarName = "GRUNTWORK_DEBUG"

// Wrapper around cli.NewApp that sets the help text printer.
func NewApp() *cli.App {
	cli.HelpPrinter = WrappedHelpPrinter
	cli.AppHelpTemplate = CLI_APP_HELP_TEMPLATE
	cli.CommandHelpTemplate = CLI_COMMAND_HELP_TEMPLATE
	cli.SubcommandHelpTemplate = CLI_APP_HELP_TEMPLATE
	app := cli.NewApp()
	return app
}

// Run the given app, handling errors, panics, and stack traces where possible
func RunApp(app *cli.App) {
	cli.OsExiter = func(exitCode int) {
		// Do nothing. We just need to override this function, as the default value calls os.Exit, which
		// kills the app (or any automated test) dead in its tracks.
	}

	defer errors.Recover(checkForErrorsAndExit)
	err := app.Run(os.Args)
	checkForErrorsAndExit(err)
}

// If there is an error, display it in the console and exit with a non-zero exit code. Otherwise, exit 0.
// Note that if the GRUNTWORK_DEBUG environment variable is set, this will print out the stack trace.
func checkForErrorsAndExit(err error) {
	exitCode := defaultSuccessExitCode
	isDebugMode := os.Getenv(debugEnvironmentVarName) != ""

	if err != nil {
		if isDebugMode {
			logging.GetLogger("").WithError(err).Error(errors.PrintErrorWithStackTrace(err))
		} else {
			fmt.Fprintf(os.Stderr, "ERROR: %s\n", errors.Unwrap(err))
		}

		errorWithExitCode, isErrorWithExitCode := err.(errors.ErrorWithExitCode)
		if isErrorWithExitCode {
			exitCode = errorWithExitCode.ExitCode
		} else {
			exitCode = defaultErrorExitCode
		}
	}

	os.Exit(exitCode)
}
