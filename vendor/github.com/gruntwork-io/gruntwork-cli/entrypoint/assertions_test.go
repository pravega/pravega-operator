package entrypoint

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

func TestStringFlagRequiredOnMissingFlag(t *testing.T) {
	t.Parallel()

	app := createSampleAppWithRequiredFlag()
	app.Action = func(cliContext *cli.Context) error {
		value, err := StringFlagRequiredE(cliContext, "the-answer-to-all-problems")
		assert.NotNil(t, err)
		assert.IsType(t, &RequiredArgsError{}, err)
		assert.Equal(t, value, "")
		return nil
	}
	args := []string{"app"}
	app.Run(args)
}

func TestStringFlagRequiredOnSetFlag(t *testing.T) {
	t.Parallel()

	app := createSampleAppWithRequiredFlag()
	app.Action = func(cliContext *cli.Context) error {
		value, err := StringFlagRequiredE(cliContext, "the-answer-to-all-problems")
		assert.Nil(t, err)
		assert.Equal(t, value, "42")
		return nil
	}
	args := []string{"app", "--the-answer-to-all-problems", "42"}
	app.Run(args)
}

func TestEnvironmentVarRequiredOnMissingEnvVar(t *testing.T) {
	value, err := EnvironmentVarRequiredE("THE_ANSWER_TO_ALL_PROBLEMS")
	assert.NotNil(t, err)
	assert.IsType(t, &RequiredArgsError{}, err)
	assert.Equal(t, value, "")
}

func TestEnvironmentVarRequiredOnSetEnvVar(t *testing.T) {
	os.Setenv("THE_ANSWER_TO_ALL_PROBLEMS", "42")
	value, err := EnvironmentVarRequiredE("THE_ANSWER_TO_ALL_PROBLEMS")
	assert.Nil(t, err)
	assert.Equal(t, value, "42")
}

func createSampleAppWithRequiredFlag() *cli.App {
	app := cli.NewApp()
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name: "the-answer-to-all-problems",
		},
	}
	return app
}
