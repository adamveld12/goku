package httpd

import (
	"fmt"
	"io"
	"strings"

	"github.com/adamveld12/gittp"
	"github.com/adamveld12/goku"
)

func newPushHandler(config goku.Configuration, backend goku.Backend) func(context gittp.HookContext, archive io.Reader) {
	logger := goku.NewLog("[push handler]")
	hostname := config.Hostname
	//dockersock := config.DockerSock
	debug := config.Debug

	return func(context gittp.HookContext, archive io.Reader) {
		cleanedBranchName := strings.TrimPrefix(context.Branch, "refs/heads/")
		logger.Tracef("Got a push to \"%v\" on the \"%v\" branch.", context.Repository, cleanedBranchName)
		context.Writeln(fmt.Sprintf("Got a push to the \"%v\" branch.", cleanedBranchName))

		p, err := newProject(archive,
			context.Repository,
			context.Commit,
			cleanedBranchName,
			hostname,
			context,
			debug)

		if err != nil {
			logger.Error(err)
			context.Writeln(fmt.Sprint("An error occurred", err.Error()))
			return
		}

		if p.Type == compose {
			// TODO implement this
			context.Writeln("Compose projects are currently not supported.")
		} else if p.Type == dockerType {
			context.Writeln("Building container")
			c, err := buildContainer(p, debug)
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
