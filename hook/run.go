package hook

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/adamveld12/goku/config"
)

func Run(path string) {
	startTime := time.Now()

	_, newRev, refName := getRevs()
	branch := strings.Replace(refName, "refs/heads/", "", 1)

	fmt.Printf("received %s:%s\ncommit: %s\n", path, branch, newRev)

	if branch != "master" {
		fmt.Println("ignoring non-master branches")
		os.Exit(128)
	}

	archiveReader, err := gitArchive(newRev)
	if err != nil {
		fmt.Println("could not read repository")
		os.Exit(128)
	}

	config.Initialize()

	config := config.Current()
	proj, err := checkout(archiveReader, path, branch, config.Domain)

	builder, err := detectProjectType(proj.Type)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(128)
	}
	if err := builder(proj); err != nil {
		fmt.Println("could not build repository")
		os.Exit(128)
	}

	fmt.Println("elapsed:", time.Since(startTime).String())
}

func detectProjectType(projType projectType) (buildFunc, error) {
	var builder buildFunc

	if projType == Composefile {
		builder = createApp
	} else if projType == Dockerfile {
		builder = createContainer
	} else {
		return nil, errors.New("Nothing to build")
	}

	return builder, nil
}
