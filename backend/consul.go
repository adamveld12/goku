package store

import (
	"errors"
	"fmt"

	. "github.com/adamveld12/goku"
	"github.com/hashicorp/consul/api"
)

const (
	gokuPrefix      = "goku"
	configKeyPrefix = gokuPrefix + "/configuration/"
	pubKeyPrefix    = gokuPrefix + "/data/keys/"
)

func init() {
	RegisterBackend("consul", consulBackendFactory)
}

func consulBackendFactory(url string) (Backend, error) {
	client, err := api.NewClient(api.DefaultConfig())

	if err != nil {
		return nil, errors.New("failed to initialize consul API")
	}

	return consulBackend{
		client,
		NewLog("[consul store]", true),
	}, nil
}

type consulBackend struct {
	*api.Client
	l Log
}

func (c consulBackend) Close() error {
	return c.Close()
}

/*
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
*/

func (c consulBackend) GetList(key string) ([][]byte, error) {
	kv := c.KV()

	pairs, _, err := kv.List(fmt.Sprintf("%s*", key), &api.QueryOptions{RequireConsistent: true})
	if err != nil {
		return nil, err
	}

	data := [][]byte{}
	for _, pair := range pairs {
		data = append(data, pair.Value)
	}

	return data, nil
}

func (c consulBackend) Put(key string, data []byte) error {
	kv := c.KV()

	p := &api.KVPair{Key: key, Value: data}
	if _, err := kv.Put(p, nil); err != nil {
		return err
	}

	return nil
}

func (c consulBackend) Delete(key string) error {
	kv := c.KV()

	if _, err := kv.Delete(key, nil); err != nil {
		return err
	}

	return nil
}

func (c consulBackend) Get(key string) ([]byte, error) {
	kv := c.KV()
	pair, _, err := kv.Get(key, nil)
	if pair == nil || err != nil {
		return nil, NilValueErr
	}

	return pair.Value, nil
}
