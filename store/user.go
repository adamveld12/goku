package store

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hashicorp/consul/api"
)

type User struct {
	PublicKey   string `json:"publickey"`
	Fingerprint string `json:"fingerprint"`
	Name        string `json:"username"`
}

func GetUsers() ([]User, error) {
	kv := client.KV()

	pairs, _, err := kv.List("goku/users/*", &api.QueryOptions{RequireConsistent: true})
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

func AddUser(u User) error {
	kv := client.KV()

	userKey := fmt.Sprintf("goku/users/%s", u.Name)
	pair, _, err := kv.Get(userKey, nil)
	if pair != nil {
		return errors.New(fmt.Sprintf("A user \"%s\" already exists.", u.Name))
	}

	if err != nil {
		return err
	}

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

func RemoveUser(name string) error {
	kv := client.KV()

	userKey := fmt.Sprintf("goku/users/%s", name)

	pair, _, err := kv.Get(userKey, nil)
	if pair == nil {
		return errors.New(fmt.Sprintf("A user \"%s\" does not exist.", name))
	}

	if err != nil {
		return err
	}

	if _, err := kv.Delete(userKey, nil); err != nil {
		return err
	}

	return nil
}

func FindUser(name string) (User, error) {
	kv := client.KV()
	pair, _, err := kv.Get(fmt.Sprintf("goku/users/%s", name), nil)

	if err != nil {
		return User{}, err
	}

	user := User{}
	if err := json.Unmarshal(pair.Value, &user); err != nil {
		return User{}, err
	}
	return user, nil
}
