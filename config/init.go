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
	"strings"

	"github.com/adamveld12/goku/log"
)

var (
	config = DefaultConfig()
	ip     = "127.0.0.1"

	configPath string
	consulUrl  string
	gitPath    string
	domainAddr string
	dockerSock string
	sshHost    string
	dataDir    string
	debug      bool
	masterOnly bool
	flagSet    *flag.FlagSet

	backend Backend
)

func init() {
	req, _ := http.NewRequest("GET", "http://ipv4.icanhazip.com/", bytes.NewBuffer([]byte{}))

	if res, err := http.DefaultClient.Do(req); err == nil {
		body, _ := ioutil.ReadAll(res.Body)
		ip = strings.Trim(string(body), "\n")
	} else {
		ip = "127.0.0.1"
	}

	flagSet = flag.NewFlagSet("goku config", flag.PanicOnError)

	flagSet.StringVar(&consulUrl, "consul", "", "url consul server")
	flagSet.StringVar(&configPath, "config", "", "path to a config.json")
	flagSet.StringVar(&gitPath, "gitpath", "repositories", "the path where remote repositories are pushed")
	flagSet.StringVar(&domainAddr, "domain", "xip.io", "domain name")
	flagSet.StringVar(&dockerSock, "dockersock", "unix:///var/run/docker.sock", "the fully qualified url to a docker daemon endpoint")
	flagSet.StringVar(&sshHost, "ssh", ":22", "ssh host and port")
	flagSet.StringVar(&dataDir, "dataDir", "./data", "the path where data is stored")
	flagSet.BoolVar(&debug, "debug", false, "enable debug mode")
	flagSet.BoolVar(&masterOnly, "masterOnly", false, "only master can be pushed")
}

func Current() Configuration {
	return config
}

// Debug returns if the app should run in debug mode
func Debug() bool {
	return debug
}

func Initialize(args []string) {
	log.Debugf(strings.Join(args, ","))

	if err := flagSet.Parse(args); err != nil {
		log.Error(err.Error())
	}

	if debug {
		ip = "127.0.0.1"
	}

	if domainAddr == "xip.io" {
		domainAddr = fmt.Sprintf("%s.xip.io", ip)
	}

	if !strings.Contains(sshHost, ":") {
		sshHost = fmt.Sprintf(":%s", sshHost)
	}

	config = DefaultConfig()

	var err error
	if consulUrl != "" {
		backend, err = ConsulBackendFactory(consulUrl)
	} else {
		backend, err = FileBackendFactory(dataDir)
	}

	if err != nil {
		log.Fatal(err.Error())
		os.Exit(128)
	}

	if configPath != "" {
		if config, err = backend.LoadConfig(configPath); err != nil {
			log.Fatal(err.Error())
			os.Exit(128)
		}
	}

	payload, _ := json.Marshal(config)
	log.Debugf("using config %s", string(payload))

}

// DefaultConfig creates a new configuration with sane defaults
func DefaultConfig() Configuration {
	return Configuration{
		RPC:               "127.0.0.1:5127",
		SSH:               sshHost,
		Hostname:          domainAddr,
		NotificationEmail: "sysadmin@example.com",
		PrivateRegistry:   "",
		GitPath:           gitPath,
		DockerSock:        dockerSock,
		DataDir:           dataDir,
		MasterOnly:        masterOnly,
		Debug:             debug,
	}
}
