package gcp

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/environment"
)

var credsEnvVars = []string{
	"GOOGLE_CREDENTIALS",
	"GOOGLE_CLOUD_KEYFILE_JSON",
	"GCLOUD_KEYFILE_JSON",
	"GOOGLE_USE_DEFAULT_CREDENTIALS",
}

var projectEnvVars = []string{
	"GOOGLE_PROJECT",
	"GOOGLE_CLOUD_PROJECT",
	"GOOGLE_CLOUD_PROJECT_ID",
	"GCLOUD_PROJECT",
	"CLOUDSDK_CORE_PROJECT",
}

var regionEnvVars = []string{
	"GOOGLE_REGION",
	"GCLOUD_REGION",
	"CLOUDSDK_COMPUTE_REGION",
}

// GetGoogleCredentialsFromEnvVar returns the Credentials for use with testing.
func GetGoogleCredentialsFromEnvVar(t *testing.T) string {
	return environment.GetFirstNonEmptyEnvVarOrEmptyString(t, credsEnvVars)
}

// GetGoogleProjectIDFromEnvVar returns the Project Id for use with testing.
func GetGoogleProjectIDFromEnvVar(t *testing.T) string {
	return environment.GetFirstNonEmptyEnvVarOrFatal(t, projectEnvVars)
}

// GetGoogleRegionFromEnvVar returns the Region for use with testing.
func GetGoogleRegionFromEnvVar(t *testing.T) string {
	return environment.GetFirstNonEmptyEnvVarOrFatal(t, regionEnvVars)
}
