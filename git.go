package goku

import (
	"archive/tar"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
)

var (
	ErrCouldNotReadRepo    = errors.New("could not read repository")
	ErrCouldNotReceiveRepo = errors.New("could not recieve repository")
	ErrCouldNotReadFile    = errors.New("could not read file header from archive")
)

type ProjectType string

const (
	Docker  = ProjectType("Dockerfile")
	Compose = ProjectType("docker.compose.yml")
	None    = ProjectType("None")
)

// Project contains meta data about the pushed repository
type Project struct {
	// Files is an array of file paths to the files in the pushed repository
	Files []string
	// Domain is the destination domain name for the pushed service once its successfully built
	Domain string
	// TargetFilePath is the target file location of the repository
	TargetFilePath string
	// Name is the name of the pushed repository as per git@<goku server>:<some/path/name>
	Name string
	// Branch is the branch that was pushed
	Branch string
	// Commit is the commit hash for this project
	Commit string
	// Archive is the tar []byte that is pushed by Git archive
	Archive []byte
	// Type is the project type. Can be either a Docker or a Compose project
	Type ProjectType

	Status io.Writer
}

func NewProject(repo io.Reader, pushedRepoName, commit, branch, domain string, status io.Writer, debug bool) (Project, error) {
	l := NewLog("\t[project processor]", debug)

	l.Trace("Processing", pushedRepoName)
	repoName := strings.Replace(pushedRepoName, ".git", "", -1)
	repoName = strings.Replace(repoName, "/", "_", -1)

	if branch != "master" {
		repoName = fmt.Sprintf("%s_%s", repoName, branch)
	}

	archive, err := ioutil.ReadAll(repo)
	if err != nil {
		l.Error("Could not open archive")
		return Project{}, ErrCouldNotReceiveRepo
	}

	proj := Project{
		Domain:  fmt.Sprintf("%s.%s", repoName, domain),
		Branch:  branch,
		Name:    repoName,
		Archive: archive,
		Commit:  commit,
		Type:    None,
		Status:  status,
	}

	arch := tar.NewReader(bytes.NewBuffer(archive))

	for {
		header, err := arch.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			return Project{}, ErrCouldNotReadRepo
		}

		if header.FileInfo().IsDir() {
			continue
		}

		fName := header.Name

		if fName == "pax_global_header" {
			continue
		}

		proj.Files = append(proj.Files, fName)

		if fName == string("CNAME") {
			data, _ := ioutil.ReadAll(arch)
			proj.Domain = string(data)
			l.Trace("Found a CNAME file, using the domain", proj.Domain)
		} else if fName == "Dockerfile" && proj.Type != Compose {
			l.Trace("Found a Dockerfile")
			proj.Type = Docker
		} else if fName == "docker.compose.yml" {
			l.Trace("Found a docker.compose.yml")
			proj.Type = Compose
		}
	}

	if proj.Type == None {
		l.Trace("Couldn't find a Dockerfile or docker.compose.yml")
		return Project{}, errors.New("This project does not have a Dockerfile or a docker.compose.yml")
	}

	return proj, nil
}
