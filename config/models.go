// config manages persisting and loading configuration
package config

import (
	"bytes"
	"encoding/json"
	"flag"
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
	ip      = "127.0.0.1"
	backend Backend

	rawBackendType = flag.String("backend", "file", "specify a backend to use")
	configPath     = flag.String("config", "", "path to a config.json or url consul server")
	gitPath        = flag.String("gitpath", "repositories", "the path where remote repositories are pushed")
	domainAddr     = flag.String("domain", "xip.io", "domain name")
	dockerSock     = flag.String("dockersock", "/var/run/docker.sock", "the url to a docker daemon")
	sshHost        = flag.String("ssh", ":22", "ssh host and port")
	dashboardHost  = flag.String("dashboard", ":80", "dashboard host and port")
	debug          = flag.Bool("debug", false, "enable debug mode")
)

func init() {
	req, _ := http.NewRequest("GET", "http://ipv4.icanhazip.com/", bytes.NewBuffer([]byte{}))

	if res, err := http.DefaultClient.Do(req); err == nil {
		body, _ := ioutil.ReadAll(res.Body)
		ip = string(body)
	}

}

// Current is a handy shortcut for getting the latest currently loaded configuration
func Current() Configuration {
	return config
}

// Debug returns if the app should run in debug mode
func Debug() bool {
	if debug == nil {
		return false
	}

	return *debug
}

func Initialize() []string {
	flag.Parse()
	if *debug {
		ip = "127.0.0.1"
	}

	if *domainAddr == "xip.io" {
		*domainAddr = fmt.Sprintf("%s.xip.io", ip)
	}

	backendType := BackendType(*rawBackendType)

	var loader BackendLoader
	if backendType == Consul {
		loader = ConsulBackendLoader
	} else if backendType == File {
		loader = FileBackendLoader
	} else {
		os.Exit(128)
	}

	var err error
	if backend, err = loader(*configPath); err != nil {
		panic(fmt.Sprintf("could not load backend\n%s", err))
	}

	config = DefaultConfig()
	loadedConfig, err := backend.LoadConfig()
	if err != nil {
		if err := backend.SaveConfig(config); err != nil {
			panic(fmt.Sprintf("could not load or save config\n%s", err.Error()))
		}
	}

	config = loadedConfig
	return flag.Args()
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
	// Domain is the domain name to use for services
	Domain string `json:"domain"`

	// AdminHost is the host/ip address to bind to for the admin dashboard
	AdminHost string `json:"admin_host"`

	// GitHost is the host/ip address to bind to for git push
	GitHost string `json:"git_host"`

	// GitPath is the file path where pushed git repositories are stored
	GitPath string `json:"git_path"`

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
		Domain:            *domainAddr,
		AdminHost:         *dashboardHost,
		GitHost:           *sshHost,
		GitPath:           *gitPath,
		NotificationEmail: "sysadmin@example.com",
		PrivateRegistry:   "",
		DockerSock:        *dockerSock,
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
