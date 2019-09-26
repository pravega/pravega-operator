package util

import (
	"log"
	"os"
)

func GetEnvOrDie(name string) string {
	value, found := os.LookupEnv(name)
	if found && value != "" {
		return value
	}

	log.Panicf("envvar %s not set or empty", name)
	return ""
}

func EnvPodNamespaceName() string {
	return GetEnvOrDie("POD_NAMESPACE")
}
