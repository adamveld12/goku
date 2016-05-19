package server

import (
	"fmt"
	"strings"

	"github.com/adamveld12/goku/config"
	"github.com/mitchellh/cli"
)

type serverCommand struct{}

func (sc serverCommand) Help() string {
	return strings.Join([]string{
		"Runs the Goku daemon, which listens for Git push and RPC calls",
		"the server will automatically create a 'git' user if needed",
		"-config: specify a config file to load",
	}, "\n")
}

func (sc serverCommand) Run(args []string) int {
	config := config.Current()

	fmt.Println(config.SSH)

	sshListen(config.SSH, config.GitPath)

	return 0
}

func (sc serverCommand) Synopsis() string {
	return "runs the Goku daemon for Git push and RPC"
}

func Command() (cli.Command, error) {
	return serverCommand{}, nil
}
