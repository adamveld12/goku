package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"sync"

	"github.com/adamveld12/goku/log"

	"github.com/fatih/color"
)

const (
	usersFilepath = "./users.json"
)

var (
	ErrNoUserFound        = errors.New("no users found")
	ErrConfigFileNotFound = errors.New("Configuration file at the specified path does not exist")
)

func FileBackendLoader(path string) (Backend, error) {
	return &fileBackend{Mutex: sync.Mutex{}, filepath: path, users: []User{}}, nil
}

type fileBackend struct {
	sync.Mutex
	filepath string
	users    []User `json:"users"`
}

func (j *fileBackend) SaveConfig(config Configuration) error {
	return ioutil.WriteFile(j.filepath, []byte(config.String()), 0)
}

func (j *fileBackend) LoadConfig() (Configuration, error) {
	data, err := ioutil.ReadFile(j.filepath)
	config := DefaultConfig()

	if err != nil && !os.IsNotExist(err) {
		color.Red("could not read config file from location %s\n%s", j.filepath, err.Error())
		return config, err
	} else {
		log.Debug("config file not found, continuing with default")
	}

	if err := json.Unmarshal(data, &config); err != nil {
		log.Fatal(fmt.Sprintf("an error occured while parsing config file at %s\n%s", j.filepath, err.Error()))
		return config, err
	}

	return config, nil
}

func (f *fileBackend) flush() error {
	f.Lock()
	defer f.Unlock()
	data, err := json.Marshal(f.users)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(f.filepath, data, 0)

}

func (f *fileBackend) Users() ([]User, error) {
	f.Lock()
	defer f.Unlock()

	return f.users, nil
}

func (f *fileBackend) SaveUser(u User) error {
	f.Lock()
	defer f.Unlock()

	f.users = append(f.users, u)

	go f.flush()

	return nil
}

func (f *fileBackend) FindUser(fingerprint string) (User, error) {
	f.Lock()
	defer f.Unlock()

	if len(f.users) == 0 {
		data, err := ioutil.ReadFile(f.filepath)
		if err != nil {
			return User{}, err
		}

		if err := json.Unmarshal(data, f.users); err != nil {
			return User{}, err
		}
	}

	for _, u := range f.users {
		if u.Fingerprint == fingerprint {
			return u, nil
		}
	}

	return User{}, ErrNoUserFound
}

func (f *fileBackend) DeleteUser(fingerprint string) error {
	f.Lock()
	defer f.Unlock()

	for idx, u := range f.users {
		if fingerprint == u.Fingerprint {
			f.users = append(f.users[:idx], f.users[idx+1:]...)
			return nil
		}
	}

	go f.flush()

	return ErrNoUserFound
}
