package hook

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	docker "github.com/fsouza/go-dockerclient"
)

const (
	Composefile = projectType("docker-compose.yml")
	Dockerfile  = projectType("Dockerfile")
	None        = projectType("None")
)

type repository struct {
	Type           projectType
	Files          []string
	TargetFilePath string
	Name           string
	Branch         string
	Archive        *[]byte
}

type projectType string

type buildFunc func(repo repository) error

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
	}

	return nil
}

func launchContainer(client *docker.Client, name string) (*docker.Container, error) {

	images, err := client.ListImages(docker.ListImagesOptions{Filter: name})

	if err != nil {
		return nil, err
	}

	targetImageId := images[0].ID
	container, err := client.CreateContainer(docker.CreateContainerOptions{
		Name:   name,
		Config: &docker.Config{Image: targetImageId},
	})

	if err != nil {
		return nil, err
	}

	if err := client.StartContainer(container.ID, &docker.HostConfig{PublishAllPorts: true}); err != nil {
		return nil, err
	}

	return container, nil
}

func Container(proj repository) error {
	fmt.Println("Dockerfile detected @", proj.TargetFilePath)

	endpoint := fmt.Sprintf("unix://%s", "/var/run/docker.sock")
	client, err := docker.NewClient(endpoint)
	if err != nil {
		return err
	}

	if err := cleanDuplicateContainer(client, proj.Name); err != nil {
		return err
	}

	if err := buildImage(client, proj.Name, *proj.Archive); err != nil {
		return err
	}

	container, err := launchContainer(client, proj.Name)
	if err != nil {
		return err
	}

	fmt.Println(container.ID)
	return nil
}