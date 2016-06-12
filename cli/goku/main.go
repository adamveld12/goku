package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/adamveld12/goku"

	_ "github.com/adamveld12/goku/httpd"
	_ "github.com/adamveld12/goku/rpcd"
)

var (
	addr       = flag.String("http", ":8080", "http address for git push and api")
	masterOnly = flag.Bool("masterOnly", true, "only allows pushing to master")
	configPath = flag.String("config", "", "path to a config.json")
	gitPath    = flag.String("gitpath", "./repositories", "path to git repositories")
	dockersock = flag.String("dockersock", "unix:///var/run/docker.sock", "path to docker daemon socket")
	host       = flag.String("host", "", "the hostname")
	debug      = flag.Bool("debug", false, "enables debug mode")
	commands   map[string]func() int
)

func main() {
	flag.Parse()

	fmt.Printf("%v\n", flag.Args())

	config, err := createConfigFromFlags()
	if err != nil {
		fmt.Println("An error occured parsing configuration inputs\n", err.Error())
		os.Exit(1)
	}

	commands = map[string]func() int{
		"server": startServer(config),
		"apps":   appsCmd(config),
		//"agent":   agent.Command,
	}

	c, ok := commands[flag.Args()[0]]
	if !ok {
		os.Exit(1)
	}

	exitStatus := c()

	os.Exit(exitStatus)
}

func startServer(config goku.Configuration) func() int {
	return func() int {
		if err := goku.StartServices(config); err != nil {
			return 1
		}
		return 0
	}
}

func createConfigFromFlags() (goku.Configuration, error) {
	cfg := goku.NewConfiguration()

	var err error
	if *configPath != "" {
		if cfg, err = goku.ConfigurationFromFile(*configPath); err != nil {
			return cfg, err
		}
	}

	// TODO testing flags taking precedence over config file
	cfg.MasterOnly = *masterOnly
	cfg.GitPath = *gitPath
	cfg.HTTP = *addr
	cfg.Debug = *debug
	cfg.DockerSock = *dockersock

	if *host != "" {
		cfg.Hostname = *host
	}

	return cfg, nil
}

func appsCmd(config goku.Configuration) func() int {
	return func() int {
		// rpc := goku.NewRPCClient(config.
		return -1
	}
}
