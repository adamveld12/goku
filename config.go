package goku

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

var (
	ip    = "127.0.0.1"
	debug = false
)

func init() {
	debug = parseDebugFromEnv()

	if detectedIP, err := detectExternalIP(); err == nil {
		ip = detectedIP
	}
}

func detectExternalIP() (string, error) {
	req, _ := http.NewRequest("GET", "http://ipv4.icanhazip.com/", nil)
	if res, err := http.DefaultClient.Do(req); err == nil {
		if body, err := ioutil.ReadAll(res.Body); err == nil {
			return strings.TrimSuffix(string(body), "\n"), nil
		}
	}

	return "", errors.New("Could not detect external ip")
}

// EnableDebug forces debug mode on
func EnableDebug() {
	debug = true
}

func parseDebugFromEnv() bool {
	return os.Getenv("DEBUG") != ""
}

// ConfigurationFromFile loads a config from a file path, returning an error if the file could not be opened or parsed
func ConfigurationFromFile(path string) (Configuration, error) {
	fs, err := os.Open(path)

	if os.IsNotExist(err) {
		return Configuration{}, err
	} else if err != nil {
		return Configuration{}, errors.New("Can't open file at path")
	}

	defer fs.Close()

	json, err := ioutil.ReadAll(fs)
	if err != nil {
		return Configuration{}, errors.New("Could not read from config file")
	}

	return ConfigurationFromJSON(json)
}

// ConfigurationFromJSON loads a config from a json string
func ConfigurationFromJSON(data []byte) (Configuration, error) {
	cfg := NewConfiguration()

	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, errors.New("Could not parse configuration from json")
	}

	return cfg, nil
}

// Configuration is a configuration struct
type Configuration struct {
	HTTP            string            `json:"http"`     // HTTP is the http bind address for git push and the dashboard API
	RPC             string            `json:"rpc"`      // RPC is the bind address for goRPC calls
	Hostname        string            `json:"hostname"` // Hostname is the host name used access apps running under Goku
	Backend         map[string]string `json:"backend"`
	PrivateRegistry string            `json:"privateRegistry"`
	GitPath         string            `json:"gitpath"`    // GitPath is the path where pushed git repositories are stored
	DockerSock      string            `json:"dockersock"` // DockerSock is the path to a docker socket. This is used to manipulate the docker daemon for running/killing containers.
	MasterOnly      bool              `json:"masterOnly"` // Only allowing pushing to the master branch
	Debug           bool              `json:"debug"`      // Enable debug printing
}

// NewConfiguration returns a Configuration with sane valid default valids
func NewConfiguration() Configuration {
	return Configuration{
		":8080",
		":5127",
		fmt.Sprintf("%s.xip.io", ip),
		map[string]string{"type": "debug"},
		"docker.io",
		"./repositories",
		"unix:///var/run/docker.sock",
		true,
		debug,
	}
}
