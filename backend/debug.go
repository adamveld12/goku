package store

import (
	"strings"
	"sync"

	. "github.com/adamveld12/goku"
)

func init() {
	RegisterBackend("debug", newDebugBackend)
}

func newDebugBackend(dir string) (Backend, error) {
	return &debugBackend{
		sync.Mutex{},
		map[string][]byte{},
	}, nil
}

type debugBackend struct {
	sync.Mutex
	store map[string][]byte
}

func (d *debugBackend) Delete(key string) error {
	d.Lock()
	defer d.Unlock()

	delete(d.store, key)
	return nil
}

func (d *debugBackend) Put(key string, data []byte) error {
	d.Lock()
	defer d.Unlock()
	d.store[key] = data
	return nil
}

func (d *debugBackend) Get(key string) ([]byte, error) {
	if v, ok := d.store[key]; ok {
		return v, nil
	}

	return nil, NilValueErr
}

func (d *debugBackend) GetList(keyPrefix string) ([][]byte, error) {
	data := [][]byte{}

	for k, v := range d.store {
		if strings.HasPrefix(k, keyPrefix) {
			data = append(data, v)
		}
	}

	return data, nil
}

func (d *debugBackend) Close() error {
	return nil
}
