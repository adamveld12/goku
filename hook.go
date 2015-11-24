package main

import (
	"archive/tar"
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

type ProjectType string

const (
	Composefile = ProjectType("docker-compose.yml")
	Dockerfile  = ProjectType("Dockerfile")
)

func buildApp(repo io.Reader, repoPath, newRev, branch string) {

	checkoutDir := strings.Replace(repoPath, "repositories", "src", 1)
	if err := checkout(repo, checkoutDir); err != nil {
		LogErr(err)
		os.Exit(128)
	}

}

func checkout(repo io.Reader, checkoutDir string) error {
	fmt.Printf("checking out repo @ %s..\n", checkoutDir)

	if _, err := os.Stat(checkoutDir); os.IsNotExist(err) {
		if err := os.MkdirAll(checkoutDir, os.ModeDir); err != nil {
			return err
		}
	}

	arch := tar.NewReader(repo)

	for {
		header, err := arch.Next()
		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		fName := header.Name
		if fName == string(Dockerfile) {
			fmt.Println("Found a dockerfile!")
			continue
		}

		if fName == string(Composefile) {
			fmt.Println("Found a compose file!")
			continue
		}

		fmt.Println(fName)
	}

	return nil
}

func runHook(path string) {

	_, newRev, refName := getRevs()
	branch := strings.Replace(refName, "refs/heads/", "", 1)

	fmt.Printf("received %s:%s\n", path, branch)
	fmt.Println(newRev)

	if branch != "master" {
		fmt.Println("ignoring non-master branches")
		os.Exit(128)
	}

	// git archive "$newrev" | "$home_dir/receiver" "$repo" "$newrev" "$user" "$fingerprint"
	gitArchive := exec.Command("git", "archive", "--format=tar", newRev)

	stdOut, err := gitArchive.StdoutPipe()
	if err != nil {
		LogError("could not read push repository")
		os.Exit(128)
	}

	if err := gitArchive.Start(); err != nil {
		LogError("could not start checkout")
		os.Exit(128)
	}

	go buildApp(stdOut, path, newRev, branch)

	if err := gitArchive.Wait(); err != nil {
		LogError("git checkout failed")
		os.Exit(128)
	}

	os.Exit(0)
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
