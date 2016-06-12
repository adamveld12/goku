package httpd

import (
	"archive/tar"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"github.com/adamveld12/goku"
)

var (
	errCouldNotReadRepo    = errors.New("could not read repository")
	errCouldNotReceiveRepo = errors.New("could not recieve repository")
	errCouldNotReadFile    = errors.New("could not read file header from archive")
)

type projectType string

const (
	dockerType = projectType("Dockerfile")
	compose    = projectType("docker.compose.yml")
	none       = projectType("None")
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
	// Type is the project type. Can be either a Docker or a Compose project
	Type projectType
	// Archive is the tar []byte that is pushed by Git archive
	Archive []byte
	// Status is for status logging to the client
	Status io.Writer
}

func newProject(gitArchive io.Reader, pushedRepoName, commit, branch, domain string, status io.Writer) (Project, error) {
	l := goku.NewLog("\t[project processor]")

	l.Trace("Processing", pushedRepoName)
	repoName := strings.Replace(pushedRepoName, ".git", "", -1)
	repoName = strings.Split(repoName, "/")[1]

	if branch != "master" {
		repoName = fmt.Sprintf("%s_%s", repoName, branch)
	}

	archive, err := ioutil.ReadAll(gitArchive)
	if err != nil {
		l.Error("Could not open archive")
		return Project{}, errCouldNotReceiveRepo
	}

	proj := Project{
		Domain:  fmt.Sprintf("%s.%s", repoName, domain),
		Branch:  branch,
		Name:    repoName,
		Commit:  commit,
		Type:    none,
		Archive: archive,
		Status:  status,
	}

	for entry := range parseTAR(archive) {
		fName := entry.Name
		proj.Files = append(proj.Files, fName)

		if entry.Name == string("CNAME") {
			data, _ := ioutil.ReadAll(entry.File)
			proj.Domain = string(data)
			l.Trace("Found a CNAME file, using the domain", proj.Domain)
		} else if fName == "Dockerfile" && proj.Type != compose {
			l.Trace("Found a Dockerfile")
			proj.Type = dockerType
		} else if fName == "docker.compose.yml" {
			l.Trace("Found a docker.compose.yml")
			proj.Type = compose
		}
	}

	if proj.Type == none {
		l.Trace("Couldn't find a Dockerfile or docker.compose.yml")
		return Project{}, errors.New("This project does not have a Dockerfile or a docker.compose.yml")
	}

	return proj, nil
}

func parseTAR(tarFile []byte) <-chan tarEntry {
	arch := tar.NewReader(bytes.NewBuffer(tarFile))
	files := make(chan tarEntry)

	go func() {
		for {
			header, err := arch.Next()
			if err == io.EOF {
				break
			}

			if header.FileInfo().IsDir() {
				continue
			}

			fName := header.Name

			if fName == "pax_global_header" {
				continue
			}

			files <- tarEntry{
				Name: fName,
				File: arch,
			}
		}

		close(files)
	}()

	return files
}

type tarEntry struct {
	Name string
	File io.Reader
}
