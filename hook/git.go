package hook

import (
	"archive/tar"
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/adamveld12/goku/log"
)

var (
	ErrCouldNotReadRepo    = errors.New("could not read repository")
	ErrCouldNotReceiveRepo = errors.New("could not recieve repository")
	ErrCouldNotReadFile    = errors.New("could not read file header from archive")
)

func gitArchive(revision string) (io.Reader, error) {
	gitArchive := exec.Command("git", "archive", "--format=tar", revision)
	stdOut, err := gitArchive.StdoutPipe()
	if err != nil {
		log.Debug("could not obtain std out pipe")
		return nil, ErrCouldNotReadRepo
	}

	if err := gitArchive.Start(); err != nil {
		log.Debug("could not start git archive")
		return nil, ErrCouldNotReadRepo
	}

	data, err := ioutil.ReadAll(stdOut)
	if err != nil {
		log.Debug("could not read from git archive out stream")
		return nil, ErrCouldNotReadRepo
	}

	if err := gitArchive.Wait(); err != nil {
		log.Debug("git archive exited with non zero exit code")
		return nil, ErrCouldNotReadRepo
	}

	return bytes.NewBuffer(data), nil
}

func checkout(repo io.Reader, repoPath, branch string) (repository, error) {
	repoName := strings.Replace(strings.TrimLeft(repoPath, "repositories/"), ".git", "", -1)

	if branch != "master" {
		repoName = fmt.Sprintf("%s_%s", repoName, branch)
	}

	archive, err := ioutil.ReadAll(repo)

	if err != nil {
		log.Debug("could not read tar of repository")
		return repository{}, ErrCouldNotReceiveRepo
	}

	proj := repository{
		Type: None,
		// TODO should be fmt.Sprintf("%.%", config.Domain, repoName)
		Domain:  fmt.Sprintf("%s.192.168.99.101.xip.io", repoName),
		Branch:  branch,
		Name:    repoName,
		Archive: &archive,
	}

	arch := tar.NewReader(bytes.NewBuffer(archive))

	for {
		header, err := arch.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			log.Debug("failed to read next file in archive")
			return repository{}, ErrCouldNotReadRepo
		}

		if header.FileInfo().IsDir() {
			continue
		}

		fName := header.Name

		if fName == "pax_global_header" {
			continue
		}

		proj.Files = append(proj.Files, fName)

		if fName == string(Dockerfile) && proj.Type != Composefile {
			proj.Type = Dockerfile
			proj.TargetFilePath = fmt.Sprintf("%s/%s", proj.Name, fName)
		}

		if fName == string(Composefile) {
			proj.Type = Composefile
			proj.TargetFilePath = fmt.Sprintf("%s/%s", proj.Name, fName)
		}

		if fName == string("CNAME") {
			data, _ := ioutil.ReadAll(arch)
			proj.Domain = string(data)
			log.Debugf("Using domain %s", proj.Domain)
		}
	}

	return proj, nil
}

func getRevs() (oldRev, newRev, refName string) {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Split(bufio.ScanWords)

	inputs := []string{}

	for scanner.Scan() {
		inputs = append(inputs, scanner.Text())
	}

	oldRev, newRev, refName = inputs[0], inputs[1], inputs[2]
	return
}
