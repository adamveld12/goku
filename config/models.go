// config manages persisting and loading configuration
package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

const (
	Consul = BackendType("consul")
	File   = BackendType("file")
)

var (
	config  = DefaultConfig()
	ip      = "0.0.0.0"
	backend Backend
)

func Initialize(rawBackendType, uri string) {
	backendType := BackendType(rawBackendType)

	var loader BackendLoader
	if backendType == Consul {
		loader = ConsulBackendLoader
	} else if backendType == File {
		loader = FileBackendLoader
	} else {
		os.Exit(128)
	}

	var err error
	if backend, err = loader(uri); err != nil {
		panic(fmt.Sprintf("could not load backend\n%s", err))
	}
}

type BackendType string
type BackendLoader func(uri string) (Backend, error)

// Backend loads/persists configuration settings for Goku
type Backend interface {
	// LoadConfig loads a Configuration with the specified URI
	LoadConfig() (Configuration, error)
	// SaveConfig saves a Configuration to the specified URI
	SaveConfig(Configuration) error
	// Users gets a list of all users saved
	Users() ([]User, error)
	// SaveUser saves a user to the backend
	SaveUser(User) error
	// ReadUser finds a user from their public key fingerprint and returns them
	FindUser(fingerprint string) (User, error)
	// DeleteUser removes a user
	DeleteUser(fingerprint string) error
}

// Configuration contains all of the configuration info used by Goku
type Configuration struct {
	// PublicIp is the public facing ip of this server
	PublicIp string `json:"public_ip"`

	// Domain is the domain name to use for services
	Domain string `json:"domain"`

	// AdminHost is the host/ip address to bind to for the admin dashboard
	AdminHost string `json:"admin_host"`

	// GitHost is the host/ip address to bind to for git push
	GitHost string `json:"git_host"`

	// NotificationEmail is an email address that gets updates for pushes
	NotificationEmail string `json:"notification_email"`

	// PrivateRegistry is the host:port for a private docker registry
	PrivateRegistry string `json:"private_registry"`

	// DockerSock is the path to the docker daemon endpoint
	DockerSock string `json:"docker_sock"`
}

func (c Configuration) String() string {
	data, err := json.Marshal(config)
	if err != nil {
		return ""
	}

	return string(data)
}

// DefaultConfig creates a new configuration with sane defaults
func DefaultConfig() Configuration {
	return Configuration{
		PublicIp:          ip,
		Domain:            fmt.Sprintf("%s.xip.io", ip),
		AdminHost:         "0.0.0.0:80",
		GitHost:           "0.0.0.0:22",
		NotificationEmail: "sysadmin@example.com",
		PrivateRegistry:   "",
		DockerSock:        "/var/run/docker.sock",
	}
}

// User represents the public key
type User struct {
	// PublicKey is this user's public key string
	PublicKey string `json:"publickey"`
	// Fingerprint is this user's public key fingerprint
	Fingerprint string `json:"fingerprint"`
	// Name is a human friendly name for this user
	Name string `json:"username"`
}

func init() {
	req, _ := http.NewRequest("GET", "http://ipv4.icanhazip.com/", bytes.NewBuffer([]byte{}))

	ip = "127.0.0.1"
	if res, err := http.DefaultClient.Do(req); err == nil {
		body, _ := ioutil.ReadAll(res.Body)
		ip = string(body)
	}
}
