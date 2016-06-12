package app

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/adamveld12/goku"
	docker "github.com/fsouza/go-dockerclient"
)

var client *docker.Client

func newDockerClient() *docker.Client {
	if client != nil {
		return client
	}

	c, err := docker.NewClient("unix:///var/run/docker.sock")

	if err != nil {
		c, err = docker.NewClientFromEnv()
	}

	if err != nil {
		panic(err)
	}

	client = c

	return client
}

func killContainer(repository, commit string) error {
	containerName := fmt.Sprintf("%s-%s", repository, commit)

	containers, err := client.ListContainers(docker.ListContainersOptions{
		Filters: map[string][]string{
			"name": []string{containerName},
		},
	})

	if err != nil || len(containers) <= 0 {
		return errors.New("Could not find a matching container")
	}

	if err := client.KillContainer(docker.KillContainerOptions{containers[0].ID, docker.SIGTERM}); err != nil {
		return fmt.Errorf("could not kill contianer %v", containerName)
	}

	return nil
}

func buildContainer(proj Project, status io.Writer) (*docker.Container, error) {
	l := goku.NewLog("\t[dockerfile builder]")

	containerImageName := fmt.Sprintf("%s-%s", proj.Commit, proj.Name)

	client := newDockerClient()

	l.Trace("Building image", containerImageName)
	status.Write([]byte("Building image...\n"))
	if err := buildImage(proj.Name, proj.Archive); err != nil {
		status.Write([]byte("Build failed\n"))
		status.Write([]byte(err.Error()))
		return nil, err
	}

	client.ListContainers(docker.ListContainersOptions{
		Filters: map[string][]string{"name": []string{containerImageName}},
	})

	l.Trace("Cleaning duplicate containers")
	status.Write([]byte("Checking for old containers...\n"))
	if err := cleanDuplicateContainer(client, proj); err != nil {
		status.Write([]byte("Container check failed -> \n"))
		status.Write([]byte(err.Error()))
		l.Error("err cleaning containers", err)
		return nil, err
	}

	l.Trace("Launching container ", proj.Name)
	status.Write([]byte("Launching container...\n"))
	container, err := launchContainer(proj.Name)
	if err != nil {
		status.Write([]byte("Launch failed\n"))
		status.Write([]byte(err.Error()))
		return nil, err
	}

	l.Trace(container.Name, " with id ", container.ID, "launched")
	return container, nil
}

type Container struct {
	Name  string
	Ports []string
	ID    string
}

func cleanDuplicateContainer(client *docker.Client, project Project) error {
	containers, err := client.ListContainers(docker.ListContainersOptions{All: true})
	if err != nil {
		return err
	}

	for _, container := range containers {
		names := container.Names
		if len(names) > 0 && project.Name == strings.TrimLeft(names[0], "/") {

			if strings.Contains(container.Status, "Up") {
				fmt.Println("stopping", project.Name)
				if err := client.KillContainer(docker.KillContainerOptions{ID: container.ID}); err != nil {
					fmt.Println("could not stop container")
					return err
				}
			}

			if err := client.RemoveContainer(docker.RemoveContainerOptions{
				ID: container.ID,
			}); err != nil {
				fmt.Println("could not remove container", err.Error())
				return err
			}

			fmt.Println("removed duplicate container")
			break
		}
	}

	return nil
}

func buildImage(name string, archive []byte) error {
	client := newDockerClient()

	if err := client.BuildImage(docker.BuildImageOptions{
		Name:         name,
		OutputStream: os.Stderr,
		InputStream:  bytes.NewBuffer(archive),
	}); err != nil {
		fmt.Println("Could not build image \n", err)
		return err
	}

	return nil
}

func launchContainer(name string) (*docker.Container, error) {

	client := newDockerClient()

	images, err := client.ListImages(docker.ListImagesOptions{Filter: name})

	if err != nil {
		return nil, err
	}

	targetImageID := images[0].ID
	container, err := client.CreateContainer(docker.CreateContainerOptions{
		Name: name,
		Config: &docker.Config{
			Image: targetImageID,
		},
	})

	if err != nil {
		return nil, err
	}

	if err := client.StartContainer(container.ID, &docker.HostConfig{PublishAllPorts: true}); err != nil {
		return nil, err
	}

	return client.InspectContainer(container.ID)
}
