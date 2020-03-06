package terraform

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/aws"
	"github.com/gruntwork-io/terratest/modules/files"
	"github.com/stretchr/testify/require"
)

func TestPlanWithNoChanges(t *testing.T) {
	t.Parallel()
	testFolder, err := files.CopyTerraformFolderToTemp("../../test/fixtures/terraform-no-error", t.Name())
	require.NoError(t, err)

	awsRegion := aws.GetRandomStableRegion(t, nil, nil)
	options := &Options{
		TerraformDir: testFolder,

		EnvVars: map[string]string{
			"AWS_DEFAULT_REGION": awsRegion,
		},
	}
	exitCode := InitAndPlan(t, options)
	require.Equal(t, DefaultSuccessExitCode, exitCode)
}

func TestPlanWithChanges(t *testing.T) {
	t.Parallel()
	testFolder, err := files.CopyTerraformFolderToTemp("../../examples/terraform-aws-example", t.Name())
	require.NoError(t, err)

	awsRegion := aws.GetRandomStableRegion(t, nil, nil)
	options := &Options{
		TerraformDir: testFolder,

		EnvVars: map[string]string{
			"AWS_DEFAULT_REGION": awsRegion,
		},
	}
	exitCode := InitAndPlan(t, options)
	require.Equal(t, TerraformPlanChangesPresentExitCode, exitCode)
}

func TestPlanWithFailure(t *testing.T) {
	t.Parallel()

	testFolder, err := files.CopyTerraformFolderToTemp("../../test/fixtures/terraform-with-plan-error", t.Name())
	require.NoError(t, err)

	options := &Options{
		TerraformDir: testFolder,
	}

	_, getExitCodeErr := InitAndPlanE(t, options)
	require.Error(t, getExitCodeErr)
}

func TestTgPlanAllNoError(t *testing.T) {
	t.Parallel()

	testFolder, err := files.CopyTerragruntFolderToTemp("../../test/fixtures/terragrunt/terragrunt-multi-plan", t.Name())
	require.NoError(t, err)

	options := &Options{
		TerraformDir:    testFolder,
		TerraformBinary: "terragrunt",
	}

	getExitCode, errExitCode := TgPlanAllExitCodeE(t, options)
	// GetExitCodeForRunCommandError was unable to determine the exit code correctly
	if errExitCode != nil {
		t.Fatal(errExitCode)
	}

	// Since PlanAllExitCodeTgE returns error codes, we want to compare against 1
	require.Equal(t, DefaultSuccessExitCode, getExitCode)
}

func TestTgPlanAllWithError(t *testing.T) {
	t.Parallel()

	testFolder, err := files.CopyTerragruntFolderToTemp("../../test/fixtures/terragrunt/terragrunt-with-plan-error", t.Name())
	require.NoError(t, err)

	options := &Options{
		TerraformDir:    testFolder,
		TerraformBinary: "terragrunt",
	}

	getExitCode, errExitCode := TgPlanAllExitCodeE(t, options)
	// GetExitCodeForRunCommandError was unable to determine the exit code correctly
	require.NoError(t, errExitCode)

	require.Equal(t, DefaultErrorExitCode, getExitCode)
}
