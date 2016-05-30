package goku

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	docker "github.com/fsouza/go-dockerclient"
)

func buildContainer(proj Project, dockersock string, debug bool) (*docker.Container, error) {
	l := NewLog("\t[dockerfile builder]", debug)

	containerImageName := fmt.Sprintf("%s-%s", proj.Branch, proj.Commit)

	l.Trace("connecting to docker daemon running @", dockersock)

	var client *docker.Client
	var err error

	if dockersock == "unix:///var/run/docker.sock" {
		l.Trace("using", dockersock)
		client, err = docker.NewClient(dockersock)
	} else {
		l.Trace("creating docker client from env")
		client, err = docker.NewClientFromEnv()
	}

	if err != nil {
		l.Error(err)
		return nil, err
	}

	l.Trace("Cleaning duplicate containers")
	proj.Status.Write([]byte("Checking for old containers -> \n"))
	if err := cleanDuplicateContainer(client, proj); err != nil {
		proj.Status.Write([]byte("Container check failed -> \n"))
		proj.Status.Write([]byte(err.Error()))
		l.Error("err cleaning containers", err)
		return nil, err
	}

	l.Trace("Building image", containerImageName)
	proj.Status.Write([]byte("Building image ->\n"))
	if err := buildImage(client, containerImageName, proj.Archive); err != nil {
		proj.Status.Write([]byte("Build failed\n"))
		proj.Status.Write([]byte(err.Error()))
		return nil, err
	}

	l.Trace("Launching container ", proj.Name)
	proj.Status.Write([]byte("Launching container ->\n"))
	container, err := launchContainer(client, containerImageName, proj.Name)
	if err != nil {
		proj.Status.Write([]byte("Launch failed\n"))
		proj.Status.Write([]byte(err.Error()))
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

func buildImage(client *docker.Client, name string, archive []byte) error {

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

func launchContainer(client *docker.Client, containerImageName, name string) (*docker.Container, error) {

	images, err := client.ListImages(docker.ListImagesOptions{Filter: containerImageName})

	if err != nil {
		return nil, err
	}

	targetImageId := images[0].ID
	container, err := client.CreateContainer(docker.CreateContainerOptions{
		Name: name,
		Config: &docker.Config{
			Image: targetImageId,
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
