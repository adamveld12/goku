package main

import (
	"os"

	"github.com/adamveld12/commando"

	"flag"
	"fmt"
)

var (
	configPath    = flag.String("config", "", "path to a config.json")
	gitPath       = flag.String("gitpath", "repositories", "the path where remote repositories are pushed")
	sshHost       = flag.String("ssh", "0.0.0.0:22", "ssh host and port")
	dashboardPort = flag.String("dashboard", "0.0.0.0:80", "dashboard host and port")
	debug         = flag.Bool("debug", false, "enable debug mode")
)

func main() {
	flag.Parse()

	InitLogging(*debug, os.Stderr)

	app := commando.New()

	app.Add("hook build", "builds a repository from the specified path and branch", hook)
	app.Add("upload-key", "uploads a public key to this repository", uploadKey)

	if err := app.Execute(flag.Args()...); err != nil {
		serverMode()
	}
}

func serverMode() {
	// go dashboardListen(*dashboardHost)
	gitListen(*sshHost, *gitPath)
}

func uploadKey(username string) {
	// read from stdin
}

func hook(path string) {
	fmt.Printf("The hook says your repo is here:\n%s", path)
}
