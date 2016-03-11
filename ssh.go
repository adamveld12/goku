package main

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/adamveld12/goku/config"
	"github.com/adamveld12/goku/log"

	"golang.org/x/crypto/ssh"
)

const (
	prereceiveHookTemplate = `#!/bin/bash
go run $GOPATH/src/github.com/adamveld12/goku/*.go -debug -domain "%s" hook %s`
	prereceiveHookPath       = `hooks/pre-receive`
	gitShellCommandErrScript = `#!/bin/sh
echo "Cannot push $REPO_NAME"
echo "$ERROR_MSG"
exit 128
`
	gitShellCommandErrPath = "~/git-shell-commands/no-interactive-login"
)

func fingerprint(pubkey ssh.PublicKey) string {
	h := md5.New()
	h.Write(pubkey.Marshal())

	rawFinger := fmt.Sprintf("%x", h.Sum(nil))

	fingerprint := ""
	for i := 0; i < len(rawFinger); i += 2 {
		fingerprint = fmt.Sprintf("%s:%s%s", fingerprint, string(rawFinger[i]), string(rawFinger[i+1]))
	}
	return strings.TrimLeft(fingerprint, ":")

}

func Listen() {
	config := config.Current()

	initGitServer(config.GitPath)

	serverConfig := ssh.ServerConfig{
		PublicKeyCallback: func(connMeta ssh.ConnMetadata, key ssh.PublicKey) (permissions *ssh.Permissions, err error) {
			username := connMeta.User()

			log.Debugf("%s - %s %s", username, key.Type(), fingerprint(key))

			if username != "git" {
				return nil, errors.New("User not found")
			}

			permissions = &ssh.Permissions{}
			err = nil

			return
		},
	}

	privateKey, err := getPrivateKey()
	if err != nil {
		log.FatalErr(err)
	}

	serverConfig.AddHostKey(privateKey)

	log.Debugf("Receiving git pushes at %s", config.GitHost)
	listener, err := net.Listen("tcp", config.GitHost)

	if err != nil {
		log.Fatal(fmt.Sprintf("failed to listen for connection:\n%s", err.Error()))
	}

	for {
		nConn, err := listener.Accept()
		log.Debug("Connection made")

		if err != nil {
			panic("failed to accept incoming connection")
		}

		go handleSSHConn(nConn, &serverConfig, config.GitPath)
	}

}

func initGitServer(gitPath string) {
	// setup repo directory
	if _, err := os.Stat(gitPath); os.IsNotExist(err) {
		log.Debugf("creating repository directory at \"%s\"", gitPath)
		if err = os.MkdirAll(gitPath, os.ModeDir); err != nil {
			log.Fatal(fmt.Sprintf("cannot create directory for git repositories at %s", gitPath))
		}
	}

	// investigate setting up https://git-scm.com/docs/git-shell
	// to show validation messages
}

func handleSSHConn(nConn net.Conn, serverConfig *ssh.ServerConfig, gitPath string) {
	_, newChans, reqChan, err := ssh.NewServerConn(nConn, serverConfig)

	if err != nil && err != io.EOF {
		log.Err(err)
		return
	}

	go ssh.DiscardRequests(reqChan)

	for newChan := range newChans {

		if newChan.ChannelType() != "session" {
			newChan.Reject(ssh.UnknownChannelType, "unknown channel type")
			continue
		}

		channel, requests, err := newChan.Accept()

		if err != nil {
			log.Errorf("Could not accept ssh connection: %s", err.Error())
		}

		go func(in <-chan *ssh.Request) {
			defer channel.Close()
			for req := range in {
				switch req.Type {
				case "exec":
					sshCommand := string(req.Payload)
					exitStatus := []byte{0, 0, 0, 0}
					if err := processPush(channel, sshCommand, gitPath); err != nil {
						log.Errorf("could not process git push\n%s", err.Error())
						exitStatus = []byte{0, 0, 0, 1}
					}
					log.Debug("git push processed successfully")
					channel.SendRequest("exit-status", false, exitStatus)
					req.Reply(true, nil)
					return
				default:
					req.Reply(false, nil)
				}
			}
		}(requests)

	}
}

func processPush(conn ssh.Channel, sshOriginalCommand, repositoryRootPath string) error {
	log.Debugf("ORIGINAL COMMAND: %s", sshOriginalCommand)

	tokens := strings.Split(sshOriginalCommand, " ")
	tokenLen := len(tokens)
	repoName := strings.Trim(tokens[tokenLen-1], "'")

	repoPath := filepath.Join(repositoryRootPath, repoName)

	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		if err := os.MkdirAll(repoPath, os.ModeDir); err != nil {
			err = errors.New("Could not make directories for repository")
			log.Err(err)
			return err
		}

		if err := createRepository(repoPath); err != nil {
			return err
		}

		config := config.Current()
		if err := createReceiveHook(repoPath, config.Domain); err != nil {
			return err
		}
	}

	if !isValidRepoName(repoName) {
		return errors.New("repository name must end with \".git\" and cannot be an empty string")
	}

	return runGitRecievePack(conn, conn.Stderr(), repoPath)
}

func runGitReceiveErr(inout io.ReadWriter, err io.Writer, repoPath, message string) error {
	os.Setenv("SSH_ORIGINAL_COMMAND", fmt.Sprintf("git-receive-pack '%s'", repoPath))
	os.Setenv("REPO_NAME", repoPath)
	os.Setenv("ERROR_MSG", message)
	cmd := exec.Command("git-shell")

	cmd.Stdin = inout
	cmd.Stderr = io.MultiWriter(err, os.Stderr)
	cmd.Stdout = inout

	if err := cmd.Run(); err != nil {
		log.Errorf("git-shell interactive failed %s", err.Error())
		return err
	}

	return nil
}

func runGitRecievePack(inout io.ReadWriter, err io.Writer, repoPath string) error {
	os.Setenv("SSH_ORIGINAL_COMMAND", fmt.Sprintf("git-receive-pack '%s'", repoPath))
	cmd := exec.Command("git-shell", "-c", fmt.Sprintf("git-receive-pack '%s'", repoPath))

	cmd.Stdin = inout
	cmd.Stderr = io.MultiWriter(err, os.Stderr)
	cmd.Stdout = inout

	if err := cmd.Run(); err != nil {
		log.Errorf("receive pack failed %s", err.Error())
		return err
	}

	return nil
}

func getPrivateKey() (ssh.Signer, error) {
	pemBytes, err := ioutil.ReadFile("/Users/Adam/.ssh/github/id_rsa")
	if err != nil {
		if pemBytes, err = generateRSAPrivateKey(); err != nil {
			return nil, errors.New("Could not generate an ssh key")
		}
	}

	pemKey, err := ssh.ParsePrivateKey(pemBytes)
	if err != nil {
		return nil, errors.New("cannot parse private key")
	}

	return pemKey, nil
}

func generateRSAPrivateKey() ([]byte, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, errors.New("cannot read private key")
	}

	priv_der := x509.MarshalPKCS1PrivateKey(privateKey)

	priv_blk := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   priv_der,
	}

	return pem.EncodeToMemory(&priv_blk), nil
}

func isValidRepoName(repoName string) bool {
	return strings.HasSuffix(repoName, ".git") && strings.Trim(repoName, " ") != ""
}

func createReceiveHook(repoPath, domain string) error {
	finalHookPath := filepath.Join(repoPath, prereceiveHookPath)
	fh, err := os.OpenFile(finalHookPath, os.O_CREATE|os.O_RDWR, 7550)

	if err != nil {
		log.Err(err)
		return err
	}

	defer fh.Close()

	prereceiveScript := fmt.Sprintf(prereceiveHookTemplate, domain, repoPath)

	if _, err := fh.WriteString(prereceiveScript); err != nil {
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
