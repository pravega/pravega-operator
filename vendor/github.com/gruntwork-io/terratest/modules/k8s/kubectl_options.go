package k8s

import "testing"

// Represents common options necessary to specify for all Kubectl calls
type KubectlOptions struct {
	ContextName string
	ConfigPath  string
	Namespace   string
	Env         map[string]string
}

func NewKubectlOptions(contextName string, configPath string) *KubectlOptions {
	return &KubectlOptions{
		ContextName: contextName,
		ConfigPath:  configPath,
	}
}

// GetConfigPath will return a sensible default if the config path is not set on the options.
func (kubectlOptions *KubectlOptions) GetConfigPath(t *testing.T) (string, error) {
	// We predeclare `err` here so that we can update `kubeConfigPath` in the if block below. Otherwise, go complains
	// saying `err` is undefined.
	var err error

	kubeConfigPath := kubectlOptions.ConfigPath
	if kubeConfigPath == "" {
		kubeConfigPath, err = GetKubeConfigPathE(t)
		if err != nil {
			return "", err
		}
	}
	return kubeConfigPath, nil
}
