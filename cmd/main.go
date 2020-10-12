package main

import (
	"log"
	"os"
	"pomerium-interview-rs-wilson/auth"
	"pomerium-interview-rs-wilson/config"
	"pomerium-interview-rs-wilson/pom"
	"pomerium-interview-rs-wilson/stats"

	"github.com/pkg/errors" // we use this error package to ensure errors are bubbled up with useful information
)

// main serves as the program entry point and error handler
func main() {
	err := run()
	if err != nil {
		log.Printf("error occurred running the pom server: %s", err)
		os.Exit(1)
	}
}

// run configures and runs our server
func run() error {
	// Get information from environment
	env, err := config.GetEnvironment()
	if err != nil {
		return errors.Wrap(err, "failed to read the environment")
	}
	env.Log()

	// Parse config file
	conf, err := config.ParseConfig(env.GetConfigFile())
	if err != nil {
		return errors.Wrap(err, "failed to read the configuration")
	}
	conf.Log()

	// Build statskeeper
	usernames := []string{}
	for _, user := range conf.Users {
		usernames = append(usernames, user.Username)
	}
	statsKeeper := stats.NewKeeper(usernames)

	// Build authprovider
	jwtAuth := auth.NewJWTAuth(conf.SharedKey)

	// Configure HTTP server
	server := pom.NewServer(env.GetListenAddress(), conf, conf, statsKeeper, jwtAuth)

	// Listen & run!
	err = server.ListenAndServe()
	if err != nil {
		return errors.Wrap(err, "failed from pom server's listen and serve")
	}

	// everything was a sucess
	return nil
}
