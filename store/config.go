// store manages persisting and loading data
package store

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/hashicorp/consul/api"
)

const configKey = "goku/configuration"

// Configuration contains all of the configuration info used by Goku
type Configuration struct {
	// PublicIp is the public ip of the server
	PublicIp string `json:"public_ip"`
	// Domain is the domain name to use for services
	Domain string `json:"domain"`
	// AdminHost is the host/ip address to bind to for the admin dashboard
	Admin string `json:"admin_host"`
	// NotificationEmail is an email address that gets updates for pushes
	NotificationEmail string `json:"notification_email"`
	// PrivateRegistry is the host:port for a private registry
	PrivateRegistry string `json:"private_registry"`
	// ConsulServer is the host:port where a consul server instance is running
	ConsulServer string `json:"consul_server"`

	// DockerSock is the path to the docker daemon endpoint
	DockerSock string `json:"docker_sock"`
}

var config = DefaultConfig()

func Config() Configuration {
	return config
}

// Initializes the configuration from a file
func LoadConfigFile(path string) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(fmt.Sprintf("could not read config file from location %s\n%s", path, err.Error()))
	}

	if err := json.Unmarshal(data, &config); err != nil {
		log.Fatal(fmt.Sprintf("an error occured while parsing config file at %s\n%s", path, err.Error()))
	}

	go ApplyConfig(config)
}

// LoadConfig fetches and returns configuration settings from storage
func LoadConfig() {
	kv := client.KV()

	pair, _, err := kv.Get(configKey, nil)

	if err != nil {
		config := DefaultConfig()
		configBytes, err := json.Marshal(config)

		if err != nil {
			panic("could not serialize default configuration")
		}

		if _, err := kv.Put(&api.KVPair{Key: configKey, Value: configBytes}, nil); err != nil {
			panic(fmt.Sprintf("could not write default config to consul\n%s\n", err.Error()))
		}

	}

	if err := json.Unmarshal(pair.Value, &config); err != nil {
		panic("could not deserialize configuration data from consul")
	}
}

// ApplyConfig applies new configuration settings over existing ones
func ApplyConfig(config Configuration) error {
	kv := client.KV()

	configBytes, err := json.Marshal(&config)
	if err != nil {
		return err
	}

	if _, err := kv.Put(&api.KVPair{Key: configKey, Value: configBytes}, nil); err != nil {
		return err
	}

	return nil
}

// DefaultConfig creates a new configuration with sane defaults
func DefaultConfig() Configuration {
	return Configuration{
		Domain:            "{public_ip}.xip.io",
		Admin:             "127.0.0.1:8080",
		NotificationEmail: "sysadmin@example.com",
		PrivateRegistry:   "",
		ConsulServer:      "0.0.0.0:8500",
		DockerSock:        "/var/run/docker.sock",
	}
}
