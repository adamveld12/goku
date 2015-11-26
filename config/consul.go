package config

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hashicorp/consul/api"
)

const (
	configKeyPrefix = "goku/configuration"
	userKeyPrefix   = "goku/users"
)

type consulBackend struct{ *api.Client }

// ConsulBackendLoader loads the consul backend storage
func ConsulBackendLoader(uri string) (Backend, error) {
	client, err := api.NewClient(api.DefaultConfig())

	if err != nil {
		return nil, errors.New("failed to initialize consul API")
	}

	return consulBackend{client}, nil
}

func (c consulBackend) LoadConfig() (Configuration, error) {
	kv := c.KV()

	pair, _, err := kv.Get(configKeyPrefix, nil)

	if err != nil {
		config := DefaultConfig()
		configBytes, err := json.Marshal(config)

		if err != nil {
			panic("could not serialize default configuration")
			return Configuration{}, err
		}

		if _, err := kv.Put(&api.KVPair{Key: configKeyPrefix, Value: configBytes}, nil); err != nil {
			panic(fmt.Sprintf("could not write default config to consul\n%s\n", err.Error()))
			return Configuration{}, err
		}
	}

	if err := json.Unmarshal(pair.Value, &config); err != nil {
		panic("could not deserialize configuration data from consul")
		return Configuration{}, err
	}
	return config, nil
}

func (c consulBackend) SaveConfig(config Configuration) error {
	kv := c.KV()

	configBytes, err := json.Marshal(&config)
	if err != nil {
		return err
	}

	if _, err := kv.Put(&api.KVPair{Key: configKeyPrefix, Value: configBytes}, nil); err != nil {
		return err
	}

	return nil
}

func (c consulBackend) Users() ([]User, error) {
	kv := c.KV()

	pairs, _, err := kv.List(fmt.Sprintf("%s*", userKeyPrefix), &api.QueryOptions{RequireConsistent: true})
	if err != nil {
		return nil, err
	}

	users := []User{}

	for _, pair := range pairs {
		user := User{}
		if err := json.Unmarshal(pair.Value, &user); err != nil {
			return nil, err
		}

		users = append(users, user)
	}

	return users, nil
}

func (c consulBackend) SaveUser(u User) error {
	kv := c.KV()

	userKey := fmt.Sprintf("%s%s", userKeyPrefix, u.Fingerprint)

	userJsonBytes, err := json.Marshal(u)
	if err != nil {
		return err
	}

	p := &api.KVPair{Key: userKey, Value: userJsonBytes}
	if _, err := kv.Put(p, nil); err != nil {
		return err
	}

	return nil
}

func (c consulBackend) DeleteUser(fingerprint string) error {
	kv := c.KV()

	userKey := fmt.Sprintf("%s%s", userKeyPrefix, fingerprint)

	pair, _, err := kv.Get(userKey, nil)
	if pair == nil {
		return errors.New(fmt.Sprintf("A user \"%s\" does not exist.", fingerprint))
	}

	if err != nil {
		return err
	}

	if _, err := kv.Delete(userKey, nil); err != nil {
		return err
	}

	return nil
}

func (c consulBackend) FindUser(fingerprint string) (User, error) {
	kv := c.KV()
	userKey := fmt.Sprintf("%s%s", userKeyPrefix, fingerprint)

	pair, _, err := kv.Get(userKey, nil)
	if pair == nil {
		return User{}, ErrNoUserFound
	}

	if err != nil {
		return User{}, err
	}

	user := User{}
	if err := json.Unmarshal(pair.Value, &user); err != nil {
		return User{}, err
	}

	return user, nil
}
