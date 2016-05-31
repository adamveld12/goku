package goku

import (
	"encoding/json"
	"errors"
	"fmt"
)

// User is a simple structure to represent a user that can interact with repositories
type User struct {
	Username     string `json:"username"`
	Email        string `json:"email"`
	PasswordHash string `json:"passwordHash"`
	PasswordSalt string `json:"passwordSalt"`
}

// UserFromJson creates a new User from a json string
func UserFromJson(userJson []byte) User {
	u := User{}
	json.Unmarshal(userJson, &u)
	return u
}

func NewUserStore(backend Backend) userStore {
	return userStore{
		backend,
	}
}

type userStore struct{ backend Backend }

func (u userStore) HandleAuth(username, password string) error {
	user, err := u.Get(username)
	if err != nil {
		return err
	}

	if password == user.PasswordHash {
		return nil
	}

	return errors.New("Unauthorized")
}

func (u userStore) Get(username string) (User, error) {
	userJson, err := u.backend.Get(createUserKey(username))
	if err != nil {
		return User{}, err
	}

	return UserFromJson(userJson), nil
}

func (u userStore) New(username, password string) (User, error) {
	return User{}, nil
}

func (u userStore) Update(user User) error {
	return nil
}

func (u userStore) Delete(username string) error {
	return u.backend.Delete(createUserKey(username))
}

func (u userStore) List() ([]User, error) {
	return nil, nil
}

func createUserKey(username string) string {
	return fmt.Sprintf("/users/%v", username)
}
