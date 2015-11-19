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

	"golang.org/x/crypto/ssh"
)

const (
	prereceiveHookTemplate = `#!/bin/bash
echo "Running hook"
go run /go/src/github.com/adamveld12/goku/*.go -debug hook %s
echo $@
exit 0
`
	prereceiveHookPath = `hooks/pre-receive`
)

func gitListen(host, gitPath string) {
	if _, err := os.Stat(gitPath); err != nil {
		LogDebugf("creating repository directory at \"%s\"", gitPath)
		if err = os.MkdirAll(gitPath, os.ModeDir); err != nil {
			LogFatal(fmt.Sprintf("cannot create directory for git repositories at %s", gitPath))
		}
	}

	serverConfig := ssh.ServerConfig{
		PublicKeyCallback: func(connMeta ssh.ConnMetadata, key ssh.PublicKey) (permissions *ssh.Permissions, err error) {
			username := connMeta.User()
			h := md5.New()
			h.Write(key.Marshal())

			LogDebugf("%s - %s %x", username, key.Type(), h.Sum(nil))

			if username != "git" {
				return nil, errors.New("User not found")
			}

			permissions = &ssh.Permissions{
				CriticalOptions: map[string]string{},
			}

			err = nil
			return
		},
	}

	privateKey, err := getPrivateKey()
	if err != nil {
		LogFatalErr(err)
	}

	serverConfig.AddHostKey(privateKey)

	listener, err := net.Listen("tcp", host)

	if err != nil {
		LogFatal(fmt.Sprintf("failed to listen for connection:\n%s", err.Error()))
	}

	for {
		nConn, err := listener.Accept()

		if err != nil {
			panic("failed to accept incoming connection")
		}

		go serveSSHConn(nConn, &serverConfig, gitPath)
	}

}

func serveSSHConn(nConn net.Conn, serverConfig *ssh.ServerConfig, gitPath string) {
	_, newChans, reqChan, err := ssh.NewServerConn(nConn, serverConfig)

	if err != nil && err != io.EOF {
		LogError(err.Error())
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
			LogErrorf("Could not accept ssh connection: %s", err.Error())
		}

		go func(in <-chan *ssh.Request) {
			defer channel.Close()
			for req := range in {
				switch req.Type {
				case "exec":
					sshCommand := string(req.Payload)
					exitStatus := []byte{0, 0, 0, 0}
					if err := processPush(channel, sshCommand, gitPath); err != nil {
						LogErrorf("could not process git push\n%s", err.Error())
						exitStatus = []byte{0, 0, 0, 1}
					}
					LogDebug("git push processed successfully")
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
	LogDebugf("ORIGINAL COMMAND: %s", sshOriginalCommand)

	tokens := strings.Split(sshOriginalCommand, " ")
	tokenLen := len(tokens)

	repoName := strings.Trim(tokens[tokenLen-1], "'")
	repoPath := filepath.Join(repositoryRootPath, repoName)

	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		if err := os.MkdirAll(repoPath, os.ModeDir); err != nil {
			err = errors.New("Could not make directories for repository")
			LogError(err.Error())
			return err
		}

		if err := createRepository(repoPath); err != nil {
			return err
		}

		if err := createReceiveHook(repoPath); err != nil {
			return err
		}
	}

	os.Setenv("SSH_ORIGINAL_COMMAND", fmt.Sprintf("git-receive-pack '%s'", repoPath))
	cmd := exec.Command("git-shell", "-c", fmt.Sprintf("git-receive-pack '%s'", repoPath))

	cmd.Stdin = conn
	cmd.Stderr = io.MultiWriter(conn.Stderr(), os.Stderr)
	cmd.Stdout = conn

	if err := cmd.Run(); err != nil {
		LogErrorf("receive pack failed %s", err.Error())
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

func createReceiveHook(repoPath string) error {
	finalHookPath := filepath.Join(repoPath, prereceiveHookPath)
	fh, err := os.OpenFile(finalHookPath, os.O_CREATE|os.O_RDWR, 7550)

	if err != nil {
		LogError(err.Error())
		return err
	}

	defer fh.Close()

	prereceiveScript := fmt.Sprintf(prereceiveHookTemplate, repoPath)

	if _, err := fh.WriteString(prereceiveScript); err != nil {
		return err
	}

	return nil
}

func createRepository(repoPath string) error {
	cmd := exec.Command("git", "init", "--bare")
	cmd.Dir = repoPath

	LogDebugf("creating a repository at %s", repoPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		LogError(err.Error())
		return errors.New("could not create remote repository")
	}

	LogDebug(string(output))

	return nil
}
