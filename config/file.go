package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
)

var (
	UnableToSaveConfigErr = errors.New("Unable to save configuration file")
	UnableToLoadConfigErr = errors.New("Unable to load configuration file")
	UnableToFlushKeysErr  = errors.New("Unable to write public keys to file")
)

func FileBackendFactory(path string) (Backend, error) {
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return nil, errors.New(fmt.Sprintf("Could not create data dir at path %s", path))
	}

	fs, err := os.Open(path + "/keys.json")
	fb := fileBackend{
		directory: path,
	}

	if err != nil && os.IsNotExist(err) {
		return fb, fb.flushKeys()
	}
	defer fs.Close()

	decoder := json.NewDecoder(fs)
	if err := decoder.Decode(&fb); err != nil {
		return nil, err
	}

	return fb, nil
}

type fileBackend struct {
	sync.Mutex
	directory string               `json:"directory"`
	keys      map[string]PublicKey `json:"keys"`
}

func (f fileBackend) LoadConfig(filepath string) (Configuration, error) {
	fs, err := os.Open(filepath)
	if err != nil {
		return Configuration{}, UnableToLoadConfigErr
	}
	defer fs.Close()

	decoder := json.NewDecoder(fs)

	configFile := Configuration{}
	if err := decoder.Decode(&configFile); err != nil {
		return Configuration{}, UnableToLoadConfigErr
	}

	return configFile, nil
}

func (f fileBackend) Keys() ([]PublicKey, error) {
	pks := []PublicKey{}
	for _, v := range f.keys {
		pks = append(pks, v)
	}

	return pks, nil
}

func (f fileBackend) AddKey(key PublicKey) error {
	f.Lock()
	defer f.Unlock()
	f.keys[key.Fingerprint] = key

	if err := f.flushKeys(); err != nil {
		return UnableToFlushKeysErr
	}

	return nil
}

func (f fileBackend) GetKey(fingerprint string) (PublicKey, error) {
	pk, ok := f.keys[fingerprint]

	if !ok {
		return PublicKey{}, PublicKeyNotFoundErr
	}

	return pk, nil
}

func (f fileBackend) DeleteKey(fingerprint string) error {
	f.Lock()
	defer f.Unlock()
	delete(f.keys, fingerprint)

	if err := f.flushKeys(); err != nil {
		return UnableToFlushKeysErr
	}

	return nil
}

func (f fileBackend) flushKeys() error {
	fs, err := os.OpenFile(f.directory+"/keys.json", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return err
	}

	defer fs.Close()

	encoder := json.NewEncoder(fs)
	if err := encoder.Encode(f); err != nil {
		return err
	}

	return nil
}
