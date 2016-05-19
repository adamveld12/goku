package server

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/adamveld12/goku/log"

	"golang.org/x/crypto/ssh"
)

func sshListen(host, path string) {

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

	privateKey, err := initGitServer(path)

	if err != nil {
		log.Fatal(err.Error())
		os.Exit(128)
	}

	serverConfig.AddHostKey(privateKey)

	listener, err := net.Listen("tcp", host)

	if err != nil {
		log.Fatal(fmt.Sprintf("failed to listen for connection:\n%s", err.Error()))
		os.Exit(128)
	}

	log.Debugf("Receiving git pushes at %s", host)
	for {
		nConn, err := listener.Accept()
		log.Debug("Connection made")

		if err != nil {
			panic("failed to accept incoming connection")
		}

		go handleSSHConn(nConn, &serverConfig, path)
	}

}

func initGitServer(gitPath string) (ssh.Signer, error) {
	// setup repo directory
	if _, err := os.Stat(gitPath); os.IsNotExist(err) {
		log.Debugf("creating repository directory at \"%s\"", gitPath)
		if err = os.MkdirAll(gitPath, os.ModeDir|os.ModePerm); err != nil {
			log.Fatal(fmt.Sprintf("cannot create directory for git repositories at %s", gitPath))
		}
	}

	privateKey, err := getPrivateKey()
	if err != nil {
		return nil, err
	}

	return privateKey, nil

	// create and return  a private key

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
						channel.Write([]byte("woops"))
						exitStatus = []byte{0, 0, 0, 1}
					} else {
						log.Debug("git push processed successfully")
					}
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
	tokens := strings.Split(sshOriginalCommand, " ")
	repoName := strings.Trim(tokens[len(tokens)-1], "'")
	repoPath := filepath.Join(repositoryRootPath, repoName)

	log.Debugf("processing push to %s", repoName)

	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		if err := os.Mkdir(repoPath, os.ModeDir|os.ModePerm); err != nil {
			err = errors.New("Could not make directories for repository")
			return err
		}

		if err := createRepository(repoPath); err != nil {
			return err
		}

		if err := createReceiveHook(repoPath); err != nil {
			return err
		}
	}

	if !isValidRepoName(repoName) {
		return errors.New("repository name must end with \".git\" and cannot be an empty string")
	}

	return runGitRecievePack(conn, conn.Stderr(), repositoryRootPath, repoName)
}

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

func getPrivateKey() (ssh.Signer, error) {
	pemBytes, err := generateRSAPrivateKey()

	if err != nil {
		return nil, errors.New("Could not generate an ssh key")
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
