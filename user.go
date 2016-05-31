package goku

import (
	"encoding/json"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUnauthorized               = errors.New("Unauthorized")
	ErrUsernameIsEmpty            = errors.New("username cannot be empty")
	ErrPasswordIsEmpty            = errors.New("password cannot be empty")
	ErrDuplicateUsername          = errors.New("a user with the specified username already exists")
	ErrCouldNotCreatePasswordHash = errors.New("could not generate password hash")
	ErrCouldNotAddUser            = errors.New("could not add user")
)

// User is a simple structure to represent a user that can interact with repositories
type User struct {
	ID           string `json:"id"`
	Username     string `json:"username"`
	Email        string `json:"email"`
	PasswordHash string `json:"passwordHash"`
}

// UserFromJson creates a new User from a json string
func UserFromJson(userJson []byte) User {
	u := User{}
	json.Unmarshal(userJson, &u)
	return u
}

func NewUserStore(backend Backend) userStore {
	return userStore{backend}
}

type userStore struct{ b Backend }

func (u userStore) HandleAuth(username, password string) error {
	user, err := u.Get(username)
	if err != nil {
		return ErrUnauthorized
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return ErrUnauthorized
	}

	return nil
}

func (u userStore) Get(username string) (User, error) {
	userJson, err := u.b.Get(createUserKey(username))
	if err != nil {
		return User{}, err
	}

	return UserFromJson(userJson), nil
}

func (u userStore) New(username, email, password string) (User, error) {
	if username == "" {
		return User{}, ErrUsernameIsEmpty
	}

	if password == "" {
		return User{}, ErrPasswordIsEmpty
	}

	if _, err := u.b.Get(createUserKey(username)); err == nil {
		return User{}, ErrDuplicateUsername
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), 15)
	if err != nil {
		return User{}, ErrCouldNotCreatePasswordHash
	}

	user := User{
		Username:     username,
		Email:        email,
		PasswordHash: string(passwordHash),
	}

	userBytes, _ := json.Marshal(u)
	if err := u.b.Put(createUserKey(username), userBytes); err != nil || len(userBytes) == 0 {
		return User{}, ErrCouldNotAddUser
	}

	return user, nil
}

func (u userStore) ChangePassword(username string, oldPassword string, newPassword string) error {
	return nil
}

func (u userStore) Delete(username string) error {
	return u.b.Delete(createUserKey(username))
}

func (u userStore) List() ([]User, error) {
	return nil, nil
}

func createUserKey(username string) string {
	return fmt.Sprintf("/users/%v", username)
}
