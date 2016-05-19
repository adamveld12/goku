package config

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/adamveld12/goku/log"
	"github.com/hashicorp/consul/api"
)

const (
	gokuPrefix      = "goku"
	configKeyPrefix = gokuPrefix + "/configuration/"
	pubKeyPrefix    = gokuPrefix + "/data/keys/"
)

func ConsulBackendFactory(url string) (Backend, error) {
	client, err := api.NewClient(api.DefaultConfig())

	if err != nil {
		return nil, errors.New("failed to initialize consul API")
	}

	return consulBackend{client}, nil
}

type consulBackend struct{ *api.Client }

func (c consulBackend) LoadConfig(path string) (Configuration, error) {
	kv := c.KV()

	pair, _, err := kv.Get(configKeyPrefix, nil)

	var configFile Configuration
	if err != nil {
		configFile = DefaultConfig()
		if err := c.SaveConfig(configFile); err != nil {
			panic("could not write default values to consul")
		}
	} else if err := json.Unmarshal(pair.Value, &configFile); err != nil {
		panic("could not deserialize configuration data from consul")
		return Configuration{}, err
	}

	return configFile, nil
}

func (c consulBackend) SaveConfig(config Configuration) error {
	kv := c.KV()

	configBytes, err := json.Marshal(&config)
	if err != nil {
		log.Err(err)
		return errors.New("could not serialize config file to json")
	}

	if _, err := kv.Put(&api.KVPair{Key: configKeyPrefix, Value: configBytes}, nil); err != nil {
		log.Err(err)
		return errors.New("could not save config file to consul")
	}

	return nil
}

func (c consulBackend) Keys() ([]PublicKey, error) {
	kv := c.KV()

	pairs, _, err := kv.List(fmt.Sprintf("%s*", pubKeyPrefix), &api.QueryOptions{RequireConsistent: true})
	if err != nil {
		return nil, err
	}

	users := []PublicKey{}

	for _, pair := range pairs {
		user := PublicKey{}
		if err := json.Unmarshal(pair.Value, &user); err != nil {
			return nil, err
		}

		users = append(users, user)
	}

	return users, nil
}

func (c consulBackend) AddKey(key PublicKey) error {
	kv := c.KV()

	keyjson, err := json.Marshal(key)
	if err != nil {
		return err
	}

	pkKey := fmt.Sprintf("%s%s", pubKeyPrefix, key.Fingerprint)
	p := &api.KVPair{Key: pkKey, Value: keyjson}
	if _, err := kv.Put(p, nil); err != nil {
		return err
	}

	return nil
}

func (c consulBackend) DeleteKey(fingerprint string) error {
	kv := c.KV()

	pkKey := fmt.Sprintf("%s%s", pubKeyPrefix, fingerprint)

	if _, err := kv.Delete(pkKey, nil); err != nil {
		return err
	}

	return nil
}

func (c consulBackend) GetKey(fingerprint string) (PublicKey, error) {
	kv := c.KV()
	pkKey := fmt.Sprintf("%s%s", pubKeyPrefix, fingerprint)

	pair, _, err := kv.Get(pkKey, nil)
	if pair == nil {
		return PublicKey{}, PublicKeyNotFoundErr
	}

	if err != nil {
		return PublicKey{}, err
	}

	pk := PublicKey{}
	if err := json.Unmarshal(pair.Value, &pk); err != nil {
		return PublicKey{}, err
	}

	return pk, nil
}
