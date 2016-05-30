# Goku

[![Build Status](https://drone.io/github.com/adamveld12/goku/status.png)](https://drone.io/github.com/adamveld12/goku/latest) [![GoDoc](https://godoc.org/github.com/adamveld12/goku?status.svg)](http://godoc.org/github.com/adamvel12/goku) [![Go Report Card](https://goreportcard.com/badge/adamveld12/goku)](https://goreportcard.com/report/adamveld12/goku)

A small, easy to install, easy to manage PaaS for hobbyists.

## Installing

Make sure the following prereqs are installed and in your PATH

- nginx - soon to be replaced internally
- git
- docker
- docker-compose

### Develop

Make sure you have Vagrant installed and then:

1. `vagrant up && vagrant ssh`
2. :sunglasses:

## How to use

### Deploying an app to Goku

1. Setup your project by adding either a `Dockerfile` or `docker-compose.yml` file in your project's root.

> This Dockerfile has to expose port 80

2. Add the remote to your repo like so: `git remote add goku http://<goku server ip/hostname>/<username>/<repository name>.git`

3. Then push: `git push goku`

You will see some validation and build output as the repository is processed.

If your repository is successfully built, Goku will publish your app at `reponame.(Goku server ip).xip.io`.


## License

MIT
