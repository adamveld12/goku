package store

import (
	"bytes"
	"errors"
	"path/filepath"

	. "github.com/adamveld12/goku"
	"github.com/boltdb/bolt"
)

func init() {
	RegisterBackend("file", newBoltBackend)
}

func newBoltBackend(dir string) (Backend, error) {
	absPath := filepath.Join(dir, "goku.db")
	db, err := bolt.Open(absPath, 0666, nil)
	if err != nil {
		return nil, errors.New("Cannot create db" + "\n" + err.Error())
	}

	return boltBackend{db, NewLog("[bolt store]", true)}, nil
}

type boltBackend struct {
	*bolt.DB
	l Log
}

func (b boltBackend) GetList(keyPrefix string) ([][]byte, error) {
	if keyPrefix == "" {
		return nil, errors.New("Key must be non empty")
	}

	keyb := []byte(keyPrefix)
	dataList := [][]byte{}
	err := b.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(keyb)

		c := bucket.Cursor()

		for k, v := c.Seek(keyb); bytes.HasPrefix(k, keyb); k, v = c.Next() {
			dataList = append(dataList, v)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return dataList, nil
}

func (b boltBackend) Delete(key string) error {
	if key == "" {
		return errors.New("Key must be non empty")
	}

	keyb := []byte(key)
	err := b.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(keyb)
		if err != nil {
			return err
		}

		if err := bucket.Delete(keyb); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return errors.New("Could not delete key")
	}

	return nil
}

func (b boltBackend) Get(key string) ([]byte, error) {
	if key == "" {
		return nil, errors.New("Key must be non empty")
	}

	keyb := []byte(key)
	var data []byte
	err := b.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(keyb)
		if bucket == nil {
			return NilValueErr
		}

		b.l.Trace("getting", key)
		data = bucket.Get(keyb)

		if len(data) == 0 {
			return NilValueErr
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return data, nil
}

func (b boltBackend) Put(key string, data []byte) error {
	if key == "" {
		return errors.New("Key must be non empty")
	}

	keyb := []byte(key)
	err := b.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(keyb)
		if err != nil {
			return err
		}

		if err := bucket.Put(keyb, data); err != nil {
			return errors.New("Could not update object at key")
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func (b boltBackend) Close() error {
	return b.DB.Close()
}
