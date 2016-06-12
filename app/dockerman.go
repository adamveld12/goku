package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/adamveld12/goku"
)

type dockerManager struct {
	goku.Log
	output io.Writer
	config goku.Configuration
	b      goku.Backend
}

// Run starts a container and adds a running container to the app list. If a duplicate container is running it will kill it
func (a *dockerManager) Run(repository, commit string) (Header, error) {
	if repository == "" {
		return Header{}, errors.New("The repository property must be defined")
	}

	if commit == "" {
		commit = "master"
	}

	header, err := a.Get(repository, commit)

	if err != nil && err != ErrAppNotExist {
		return Header{}, err
	} else if err != ErrAppNotExist {
		// we're overwriting the old one, so we need to kill it
		if err := a.Kill(header.Repository, header.Commit); err != nil {
			return Header{}, err
		}
	}

	config := a.config
	fullPath := filepath.Join(config.GitPath, repository)
	repositoryName := strings.Trim(filepath.Base(repository), ".git")
	p, err := newProject(fullPath, repositoryName, commit, config.Hostname)
	if err != nil {
		return Header{}, err
	}

	if err := buildImage(fmt.Sprintf("%v-%v", repositoryName, commit), p.Archive); err != nil {
		return Header{}, err
	}

	c, err := buildContainer(p, a.output)
	if err != nil {
		return Header{}, err
	}

	header = Header{
		Name:        p.Name,
		Repository:  repositoryName,
		Commit:      p.Commit,
		URL:         p.Domain,
		ContainerID: c.ID,
		Status:      c.State.String(),
		StartTime:   c.Created,
	}

	jsonBytes, _ := json.Marshal(header)
	if err := a.b.Put(fmt.Sprintf("goku/apps/%v", p.Name), jsonBytes); err != nil {
		return Header{}, err
	}

	return header, nil
}

func (a *dockerManager) List(filter string) ([]Header, error) {
	return []Header{}, nil
}

func (a *dockerManager) Get(repository, commit string) (Header, error) {
	return Header{}, nil
}

// Kill stops the container(s) running under the specified application
func (a *dockerManager) Kill(repository, commit string) error {
	_, err := a.Get(repository, commit)
	if err != nil {
		return err
	}

	return nil
}
