package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/adamveld12/goku"
	"github.com/adamveld12/goku/httpd"
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
		backend, _ := goku.NewBackend("debug", "")

		sv, err := httpd.New(config, backend)
		if err != nil {
			return 1
		}

		if err := sv.Start(); err != nil {
			log.Println(err.Error())
			return 1
		}

		sigs := make(chan os.Signal, 2)
		signal.Notify(sigs, os.Interrupt)
		<-sigs
		signal.Stop(sigs)

		fmt.Println("Stopping http server...")
		if err := sv.Stop(); err != nil {
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
