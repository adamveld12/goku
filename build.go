package goku

import (
	"fmt"
	"io"
	"strings"

	"github.com/adamveld12/gittp"
)

func NewPushHandler(config Configuration) func(context gittp.HookContext, archive io.Reader) {
	logger := NewLog("[push handler]", config.Debug)
	return func(context gittp.HookContext, archive io.Reader) {
		cleanedBranchName := strings.TrimPrefix(context.Branch, "refs/heads/")
		logger.Tracef("Got a push to \"%v\" on the \"%v\" branch.", context.Repository, cleanedBranchName)
		context.Writeln(fmt.Sprintf("Got a push to the \"%v\" branch.", cleanedBranchName))

		p, err := NewProject(archive,
			context.Repository,
			context.Commit,
			cleanedBranchName,
			config.Hostname,
			context,
			config.Debug)

		if err != nil {
			logger.Error(err)
			context.Writeln(fmt.Sprint("An error occurred", err.Error()))
			return
		}

		if p.Type == Compose {
			// TODO implement this
		} else if p.Type == Docker {

			context.Writeln("Building container")
			c, err := buildContainer(p, config.DockerSock, config.Debug)
			if err != nil {
				logger.Error(err)
				context.Writeln("Build failed")
				return
			}

			if err := publish(p, c); err != nil {
				logger.Error(err)
				context.Writeln("Could not publish")
				return
			}
		}

		logger.Trace("Push succeeded")
		context.Writeln("Push succeeded")
		context.Writeln("your app is running at http://" + p.Domain)
	}
}

func handlePush(context gittp.HookContext, p Project) error {
	return nil
}
