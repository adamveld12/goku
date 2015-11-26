package main

import (
	"os"

	"github.com/adamveld12/commando"
	"github.com/adamveld12/goku/hook"

	"flag"
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

	app.Add("generate-config", "generates a commented config.json with sane defaults at the specified location", genConfig)
	app.Add("hook build", "builds a repository from the specified path and branch", hook.Run)

	if err := app.Execute(flag.Args()...); err != nil {
		serverMode()
	}
}

func genConfig(path string) {

}

func serverMode() {
	// go dashboardListen(*dashboardHost)
	gitListen(*sshHost, *gitPath)
}
