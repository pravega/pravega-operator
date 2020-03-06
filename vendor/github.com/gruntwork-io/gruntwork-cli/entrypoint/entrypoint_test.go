package entrypoint

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
	"testing"
)

func TestEntrypointNewAppWrapsAppHelpPrinter(t *testing.T) {
	app := createSampleApp()
	fakeStdout := bytes.NewBufferString("")
	app.Writer = fakeStdout
	args := []string{"houston", "help"}
	err := app.Run(args)
	assert.Nil(t, err)
	assert.Equal(t, fakeStdout.String(), EXPECTED_APP_HELP_OUT)
}

func TestEntrypointNewAppWrapsCommandHelpPrinter(t *testing.T) {
	app := createSampleApp()
	fakeStdout := bytes.NewBufferString("")
	app.Writer = fakeStdout
	args := []string{"houston", "help", "exec"}
	err := app.Run(args)
	assert.Nil(t, err)
	assert.Equal(t, fakeStdout.String(), EXPECTED_EXEC_CMD_HELP_OUT)
}

func TestEntrypointNewAppHelpPrinterHonorsLineWidthVar(t *testing.T) {
	HelpTextLineWidth = 120
	app := createSampleApp()
	fakeStdout := bytes.NewBufferString("")
	app.Writer = fakeStdout
	args := []string{"houston", "help"}
	err := app.Run(args)
	assert.Nil(t, err)
	assert.Equal(t, fakeStdout.String(), EXPECTED_APP_HELP_OUT_120_LINES)
}

func TestEntrypointNewAppCommandHelpPrinterHonorsLineWidthVar(t *testing.T) {
	HelpTextLineWidth = 120
	app := createSampleApp()
	fakeStdout := bytes.NewBufferString("")
	app.Writer = fakeStdout
	args := []string{"houston", "help", "exec"}
	err := app.Run(args)
	assert.Nil(t, err)
	assert.Equal(t, fakeStdout.String(), EXPECTED_EXEC_CMD_HELP_OUT_120_LINES)
}

func noop(c *cli.Context) error { return nil }

func createSampleApp() *cli.App {
	app := NewApp()
	app.Name = "houston"
	app.HelpName = "houston"
	app.Version = "v0.0.6"
	app.Description = `A CLI tool for interacting with Gruntwork Houston that you can use to authenticate to AWS on the CLI and to SSH to your EC2 Instances.`

	configFlag := cli.StringFlag{
		Name:  "config, c",
		Value: "~/.houston/houston.yml",
		Usage: "The configuration file for houston",
	}

	portFlag := cli.IntFlag{
		Name:  "port",
		Value: 44444,
		Usage: "The TCP port the http server is running on",
	}

	app.Commands = []cli.Command{
		{
			Name:      "exec",
			Usage:     "Execute a command with temporary AWS credentials obtained by logging into Gruntwork Houston",
			UsageText: "houston exec [options] <profile> -- <command>",
			Description: `The exec command makes it easier to use CLI tools that need AWS credentials, such as aws, terraform, and packer. Here's how it works:

   1. The first time you run this command for a <profile>, it will open your web browser and have you login to your Identity Provider (i.e., Google, ADFS, Okta). 
   2. After login, the Identity Provider will redirect you to the Gruntwork Houston web console, where you will pick the AWS IAM Role you want to use. 
   3. Gruntwork Houston will fetch temporary AWS credentials for this IAM Role and POST them back to houston CLI running on your computer.
   4. The houston CLI will set those credentials as the appropriate environment variables and execute <command>.
   5. The houston CLI will cache those credentials in memory, so all subsequent commands will execute without going through the login flow (until those credentials expire).

Examples:

   houston exec dev -- aws s3 ls
   houston exec prod -- terraform apply
   houston exec stage -- packer build server.json`,
			Action: noop,
			Flags:  []cli.Flag{configFlag, portFlag},
		},
		{
			Name:      "ssh",
			Usage:     "Connect to an EC2 instance via SSH with your public key.",
			UsageText: "houston ssh [OPTIONS] <username>@<ip>",
			Description: `Pre-requisites for using this command:

   1. Install ssh-grunt on your EC2 Instances.
   2. Configure each EC2 Instance to talk to your Gruntwork Houston deployment and to grant access to users with certain SSH Roles.
   3. Have an admin grant you those SSH Roles in your Identity Provider (e.g., Google, ADFS, Okta).
   4. Upload your public SSH key into Gruntwork Houston using its web console.

Once all the above is taken care of, here is how the houston ssh command works:

   1. If you haven't logged into the Gruntwork Houston web console recently, it will pop open your web browser and ask you to login.
   2. Gruntwork Houston will cache the SSH Roles you have for a short time period. The need to update this cache is why the houston ssh command exists!
   3. ssh-grunt runs a cron job on your EC2 Instances that creates local OS users for Houston users with specific SSH Roles.
   4. The houston CLI on your computer will execute your native ssh command to connect to the EC2 Instance.
   5. When the SSH request comes into the EC2 Instance, ssh-grunt fetches your public SSH key from Gruntwork Houston and gives it to the SSH daemon for verification. 

Examples:

   houston ssh grunt@11.22.33.44`,
			Action: noop,
			Flags:  []cli.Flag{configFlag, portFlag},
		},
		{
			Name:        "configure",
			Usage:       "Configure houston CLI options.",
			UsageText:   "houston configure [options]",
			Description: `The configure command can be used to setup or update the houston configuration file. When you run this command with no arguments, it will prompt you with the minimum required options for getting the CLI up and running. The prompt will include the current value as a default if the configuration file exists.`,
			Action:      noop,
			Flags:       []cli.Flag{configFlag},
		},
	}
	return app
}

const EXPECTED_APP_HELP_OUT = `Usage: houston [--help] [--version] command [options] [args]

A CLI tool for interacting with Gruntwork Houston that you can use to
authenticate to AWS on the CLI and to SSH to your EC2 Instances.

Commands:

   exec       Execute a command with temporary AWS credentials obtained by
              logging into Gruntwork Houston
   ssh        Connect to an EC2 instance via SSH with your public key.
   configure  Configure houston CLI options.
   help, h    Shows a list of commands or help for one command

`

const EXPECTED_APP_HELP_OUT_120_LINES = `Usage: houston [--help] [--version] command [options] [args]

A CLI tool for interacting with Gruntwork Houston that you can use to authenticate to AWS on the CLI and to SSH to your
EC2 Instances.

Commands:

   exec       Execute a command with temporary AWS credentials obtained by logging into Gruntwork Houston
   ssh        Connect to an EC2 instance via SSH with your public key.
   configure  Configure houston CLI options.
   help, h    Shows a list of commands or help for one command

`

const EXPECTED_EXEC_CMD_HELP_OUT = `Usage: houston exec [options] <profile> -- <command>

The exec command makes it easier to use CLI tools that need AWS credentials,
such as aws, terraform, and packer. Here's how it works:

   1. The first time you run this command for a <profile>, it will open your web
   browser and have you login to your Identity Provider (i.e., Google, ADFS,
   Okta).
   2. After login, the Identity Provider will redirect you to the Gruntwork
   Houston web console, where you will pick the AWS IAM Role you want to use.
   3. Gruntwork Houston will fetch temporary AWS credentials for this IAM Role
   and POST them back to houston CLI running on your computer.
   4. The houston CLI will set those credentials as the appropriate environment
   variables and execute <command>.
   5. The houston CLI will cache those credentials in memory, so all subsequent
   commands will execute without going through the login flow (until those
   credentials expire).

Examples:

   houston exec dev -- aws s3 ls
   houston exec prod -- terraform apply
   houston exec stage -- packer build server.json

Options:

   --config value, -c value  The configuration file for houston (default:
                             "~/.houston/houston.yml")
   --port value              The TCP port the http server is running on (default:
                             44444)

`

const EXPECTED_EXEC_CMD_HELP_OUT_120_LINES = `Usage: houston exec [options] <profile> -- <command>

The exec command makes it easier to use CLI tools that need AWS credentials, such as aws, terraform, and packer. Here's
how it works:

   1. The first time you run this command for a <profile>, it will open your web browser and have you login to your
   Identity Provider (i.e., Google, ADFS, Okta).
   2. After login, the Identity Provider will redirect you to the Gruntwork Houston web console, where you will pick the
   AWS IAM Role you want to use.
   3. Gruntwork Houston will fetch temporary AWS credentials for this IAM Role and POST them back to houston CLI running
   on your computer.
   4. The houston CLI will set those credentials as the appropriate environment variables and execute <command>.
   5. The houston CLI will cache those credentials in memory, so all subsequent commands will execute without going
   through the login flow (until those credentials expire).

Examples:

   houston exec dev -- aws s3 ls
   houston exec prod -- terraform apply
   houston exec stage -- packer build server.json

Options:

   --config value, -c value  The configuration file for houston (default: "~/.houston/houston.yml")
   --port value              The TCP port the http server is running on (default: 44444)

`
