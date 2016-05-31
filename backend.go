package goku

import "errors"

var (
	activebackendType string
	activeBackend     Backend
	stores            = map[string]BackendFactory{}
	NilValueErr       = errors.New("No value found for specified key")
)

// BackendFactory is a func that can initialize and return a Backend implementation. This object is cached for later use
type BackendFactory func(string) (Backend, error)

// Backend is an interface for a storage backend
type Backend interface {
	GetList(key string) ([][]byte, error)
	Get(key string) ([]byte, error)
	Put(key string, value []byte) error
	Delete(key string) error
	Close() error
}

// RegisterBackend registers a backend
func RegisterBackend(backendType string, bf BackendFactory) {
	if backendType == "" {
		panic("backendType cannot be empty")
	}

	stores[backendType] = bf
}

// NewBackend returns a new Backend implementation based on the backendType string. Uri can be a connection string or a file/directory path depending on the implementation being used
func NewBackend(backendType, uri string) (Backend, error) {
	if activeBackend != nil {
		if activebackendType != backendType {
			return nil, errors.New("A backend with a different type has already been created")
		}

		return activeBackend, nil
	}

	backendFactory, ok := stores[backendType]

	if !ok {
		return nil, errors.New("backend type not registered")
	}

	backend, err := backendFactory(uri)
	if err != nil {
		return nil, err
	}

	activeBackend = backend
	activebackendType = backendType

	return backend, nil
}
