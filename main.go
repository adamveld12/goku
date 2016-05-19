package main

import (
	"os"

	"github.com/adamveld12/goku/config"
	"github.com/adamveld12/goku/log"

	"github.com/adamveld12/goku/server"
	"github.com/mitchellh/cli"
)

func main() {
	config.Initialize(os.Args[2:])
	log.Initialize(config.Debug(), os.Stderr)

	c := cli.NewCLI("goku", "1.0.0")
	c.Args = os.Args[1:]

	c.Commands = map[string]cli.CommandFactory{
		"server": server.Command,
		//"keys":   keys.Command,
		//"hook": hook.Command(),
	}

	exitStatus, err := c.Run()
	if err != nil {
		log.Println(err.Error())
	}

	os.Exit(exitStatus)
}
