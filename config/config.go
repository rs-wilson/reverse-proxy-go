package config

import (
	"encoding/json"
	"io/ioutil"
	"log"

	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

// ParseConfig parses a json file into retrievable information
func ParseConfig(file string) (*PomConfig, error) {
	//read file
	fileData, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read config file")
	}

	// read JSON object
	config := &PomConfig{}
	err = json.Unmarshal(fileData, config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal json config file bytes")
	}

	// convert to helper maps
	config.passHashes = make(map[string]string)
	for _, user := range config.Users {
		config.passHashes[user.Username] = user.PassHash
	}

	config.destinations = make(map[string]string)
	config.userGate = make(map[string][]string)
	for _, route := range config.Routes {
		config.destinations[route.Name] = route.Destination
		config.userGate[route.Name] = route.AllowedUsers
	}

	return config, nil
}

// PomConfig doubles as a config reader and an authentication checker (this isn't perfect design but it's easy for now)
type PomConfig struct {
	SharedKey string  `json:"shared_key"`
	Users     []User  `json:"users"`
	Routes    []Route `json:"allowed_routes"`

	passHashes   map[string]string
	destinations map[string]string
	userGate     map[string][]string
}

type User struct {
	Username string `json:"username"`
	PassHash string `json:"password_hash"`
}

type Route struct {
	Name         string   `json:"name"`
	Destination  string   `json:"destination"`
	AllowedUsers []string `json:"users"`
}

func (me *PomConfig) Log() {
	log.Printf("Config:\n")
	log.Printf("\tShared Key: %s", me.SharedKey)
	log.Printf("\tUsers: %d", len(me.Users))
	log.Printf("\tRoutes: %d", len(me.Routes))
}

// CheckUsername returns true if the username is known to the configuration
func (me *PomConfig) CheckUsername(username string) bool {
	_, ok := me.passHashes[username]
	if !ok {
		return false
	}
	return true
}

// CheckPassword returns true if the user/pass combination matches our internal data
func (me *PomConfig) CheckPassword(username string, password string) bool {
	hashed, ok := me.passHashes[username]
	if !ok {
		return false
	}
	err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password))
	if err != nil {
		return false
	}
	return true
}

// GetAddress returns the destination url for a named target, or false is none is found
func (me *PomConfig) GetAddress(target string) (string, bool) {
	addr, ok := me.destinations[target]
	return addr, ok
}

// IsIsUserAllowed returns true if the user is allowed to be redirected to the given url
func (me *PomConfig) IsUserAllowed(target string, username string) bool {
	users, ok := me.userGate[target]
	if !ok {
		return false
	}
	for _, user := range users {
		if user == username {
			return true
		}
	}
	return false //the user was not found in our list of allowed users
}
