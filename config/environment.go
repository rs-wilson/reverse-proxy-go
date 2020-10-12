package config

import (
	"fmt"
	"log"
	"os"

	"github.com/pkg/errors"
)

// This file contains helper functions for retrieving the environment in which the server is run

const portEnv = "POM_PORT"
const configEnv = "POM_CONFIG"

// GetEnvironment returns the expected environment variables and their values
func GetEnvironment() (*PomEnv, error) {
	env := &PomEnv{
		envs: make(map[string]string),
	}

	// Listen Port
	portVal, err := getEnvVar(portEnv)
	if err != nil {
		return nil, err
	}
	env.envs[portEnv] = portVal

	// Readable config
	configVal, err := getEnvVar(configEnv)
	if err != nil {
		return nil, err
	}
	env.envs[configEnv] = configVal

	return env, nil
}

func getEnvVar(key string) (string, error) {
	value, ok := os.LookupEnv(key)
	if !ok {
		return "", errors.Errorf("failed to get value for env var %s", key)
	}
	return value, nil
}

type PomEnv struct {
	envs map[string]string
}

func (me *PomEnv) Log() {
	log.Printf("Environment:\n")
	for k, v := range me.envs {
		log.Printf("\t%s: %s", k, v)
	}
}

func (me *PomEnv) GetListenAddress() string {
	return fmt.Sprintf(":%s", me.envs[portEnv])
}

func (me *PomEnv) GetConfigFile() string {
	return me.envs[configEnv]
}
