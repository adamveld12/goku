package server

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/adamveld12/goku/config"
	"github.com/adamveld12/goku/log"
)

var (
	prereceiveHookPath = `hooks/pre-receive`

	gitShellCommandErrScript = `#!/bin/sh
echo "Cannot push $REPO_NAME"
echo "$ERROR_MSG"
exit 128
`
	gitShellCommandErrPath = "~/git-shell-commands/no-interactive-login"
)

func getPrereceiveHook(args string) string {
	config := config.Current()

	prereceiveTempl := []string{
		"#!/bin/bash\necho git push $REPO_NAME successful",
	}

	buildCommand := "goku build %s"
	if config.Debug {
		buildCommand = "go run $GOPATH/src/github.com/adamveld12/goku/*.go build -debug %s"
	}

	return fmt.Sprintf(
		strings.Join(append(prereceiveTempl, buildCommand), ";\n"), args)
}

func isValidRepoName(repoName string) bool {
	return strings.HasSuffix(repoName, ".git") && strings.Trim(repoName, " ") != ""
}

func createReceiveHook(repoPath string) error {
	finalHookPath := filepath.Join(repoPath, prereceiveHookPath)
	fh, err := os.OpenFile(finalHookPath, os.O_CREATE|os.O_RDWR, 7550)

	if err != nil {
		log.Err(err)
		return err
	}

	defer fh.Close()

	if _, err := fh.WriteString(getPrereceiveHook(repoPath)); err != nil {
		return err
	}

	return nil
}

func createRepository(repoPath string) error {
	cmd := exec.Command("git", "init", "--bare")
	cmd.Dir = repoPath

	log.Debugf("creating a repository at %s", repoPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Err(err)
		return errors.New("could not create remote repository")
	}

	log.Debug(string(output))

	return nil
}

func runGitRecievePack(inout io.ReadWriter, err io.Writer, repoRoot, repoName string) error {
	originalCommand := fmt.Sprintf("git-receive-pack '%s'", repoName)

	cmd := exec.Command("git-shell", "-c", originalCommand)
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("SSH_ORIGINAL_COMMAND=%s", originalCommand),
		fmt.Sprintf("REPO_NAME=%s", repoName))
	cmd.Dir = repoRoot

	cmd.Stdin = inout
	cmd.Stderr = io.MultiWriter(err, os.Stderr)
	cmd.Stdout = inout

	if err := cmd.Run(); err != nil {
		log.Errorf("receive pack failed %s", err.Error())
		return err
	}

	return nil
}
