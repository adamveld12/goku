package main

import (
	"os"

	"github.com/adamveld12/commando"
	"github.com/adamveld12/goku/config"
	"github.com/adamveld12/goku/hook"
	"github.com/adamveld12/goku/log"
)

func main() {
	args := config.Initialize()

	log.Initialize(config.Debug(), os.Stderr)

	app := commando.New()

	app.Add("generate-config", "generates a commented config.json with sane defaults at the specified location", genConfig)
	app.Add("hook build", "builds a repository from the specified path", hook.Run)

	if err := app.Execute(args...); err != nil {
		serverMode()
	}
}

func genConfig(path string) {}

func serverMode() {
	// go dashboardListen(*dashboardHost)
	Listen()
}
