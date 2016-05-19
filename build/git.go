package build

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
	"time"

	"github.com/adamveld12/goku/config"
	"github.com/adamveld12/goku/log"
	"github.com/mitchellh/multistep"
)

var (
	ErrCouldNotReadRepo    = errors.New("could not read repository")
	ErrCouldNotReceiveRepo = errors.New("could not recieve repository")
	ErrCouldNotReadFile    = errors.New("could not read file header from archive")
)

type processPushStep struct{}

func (irs processPushStep) Run(state multistep.StateBag) multistep.StepAction {
	_, newRev, refName := getRevs()
	branch := strings.Replace(refName, "refs/heads/", "", 1)
	repoName := state.Get("repoName").(string)

	configuration := config.Current()

	if branch != "master" && configuration.MasterOnly {
		fmt.Println("ignoring non-master branches")
		state.Put("exit", 128)
		return multistep.ActionHalt
	}

	archiveReader, err := gitArchive(newRev)
	if err != nil {
		fmt.Println("could not read repository")
		state.Put("exit", 128)
		return multistep.ActionHalt
	}

	project, err := checkout(archiveReader,
		repoName,
		configuration.GitPath,
		branch,
		configuration.Hostname)

	if err != nil {
		fmt.Println("could not checkout project at %s", branch)
		state.Put("exit", 128)
		return multistep.ActionHalt
	}

	state.Put("startTime", time.Now())
	state.Put("newRev", newRev)
	state.Put("branch", refName)
	state.Put("project", project)

	return multistep.ActionContinue
}

func (ppS processPushStep) Cleanup(state multistep.StateBag) {}

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
	// Archive is the tar []byte that is pushed by Git archive
	Archive *[]byte
}

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

func getRevs() (string, string, string) {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Split(bufio.ScanWords)

	inputs := []string{}

	for scanner.Scan() {
		inputs = append(inputs, scanner.Text())
	}

	return inputs[0], inputs[1], inputs[2]
}

func checkout(repo io.Reader, pushedRepoName, repoPath, branch, domain string) (Project, error) {
	repoName := strings.Replace(pushedRepoName, ".git", "", -1)

	if branch != "master" {
		repoName = fmt.Sprintf("%s_%s", repoName, branch)
	}

	archive, err := ioutil.ReadAll(repo)

	if err != nil {
		log.Debug("could not read tar of repository")
		return Project{}, ErrCouldNotReceiveRepo
	}

	proj := Project{
		Domain:  fmt.Sprintf("%s.%s", repoName, domain),
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
			log.Debugf("Using domain %s", proj.Domain)
		}
	}

	return proj, nil
}
