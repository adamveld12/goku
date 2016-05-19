package build

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/adamveld12/goku/config"
	"github.com/adamveld12/goku/log"
	docker "github.com/fsouza/go-dockerclient"
)

type dockerfileBuilder struct{}

func (dfb dockerfileBuilder) IsMatch(filename string) bool {
	return filename == "Dockerfile"
}

func (dfb dockerfileBuilder) Build(proj Project, endpoint string) (string, error) {
	fmt.Println("Dockerfile detected @", proj.TargetFilePath)

	config := config.Current()
	endpoint = config.DockerSock
	client, err := docker.NewClient(endpoint)
	if err != nil {
		log.DebugErr(err)
		return "", err
	}

	if err := cleanDuplicateContainer(client, proj.Name); err != nil {
		log.DebugErr(err)
		return "", err
	}

	if err := buildImage(client, proj.Name, *proj.Archive); err != nil {
		log.Debugf("could not build image\n%s", err)
		return "", err
	}

	container, err := launchContainer(client, proj.Name)
	if err != nil {
		log.Debugf("could not launch container\n%s", err.Error())
		return "", err
	}

	fmt.Println(container.ID, ":", container.Name)
	return container.ID, nil
}

type container struct {
	Name  string
	Ports []string
	ID    string
}

func killContainer(client *docker.Client, containerID string) error {
	if err := client.KillContainer(docker.KillContainerOptions{ID: containerID}); err != nil {
		return errors.New("Could not kill container")
	}
	return nil
}

func cleanDuplicateContainer(client *docker.Client, name string) error {

	containers, err := client.ListContainers(docker.ListContainersOptions{All: true})
	if err != nil {
		return err
	}

	for _, container := range containers {
		names := container.Names
		if len(names) > 0 && name == strings.TrimLeft(names[0], "/") {

			if strings.Contains(container.Status, "Up") {
				fmt.Println("stopping", name)
				if err := client.KillContainer(docker.KillContainerOptions{ID: container.ID}); err != nil {
					fmt.Println("could not stop container")
					log.DebugErr(err)
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
		OutputStream: os.Stdout,
		InputStream:  bytes.NewBuffer(archive),
	}); err != nil {
		fmt.Println("Could not build image \n", err)
		return err
	}

	return nil
}

func launchContainer(client *docker.Client, name string) (*docker.Container, error) {

	images, err := client.ListImages(docker.ListImagesOptions{Filter: name})

	if err != nil {
		log.DebugErr(err)
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
		log.DebugErr(err)
		return nil, err
	}

	if err := client.StartContainer(container.ID, &docker.HostConfig{PublishAllPorts: true}); err != nil {
		log.DebugErr(err)
		return nil, err
	}

	return client.InspectContainer(container.ID)
}
