# Goku

[![Build Status](https://drone.io/github.com/adamveld12/goku/status.png)](https://drone.io/github.com/adamveld12/goku/latest) [![GoDoc](https://godoc.org/github.com/adamveld12/goku?status.svg)](http://godoc.org/github.com/adamvel12/goku) [![Go Report Card](https://goreportcard.com/badge/adamveld12/goku)](https://goreportcard.com/report/adamveld12/goku)

A small, easy to install, easy to manage PaaS for hobbyists.

## Installing

Make sure the following prereqs are installed and in your PATH

- nginx - soon to be replaced internally
- git
- docker
- docker-compose


## How to use

### Configuring Goku

Goku can be configured by a .json file that can be loaded like so:

`goku server -config /path/to/config.json`

The config file (with the system defaults) is detailed below:

```js
{
  "http": "0.0.0.0:80", // The interface:port that the http server listens on
  "rpc": "127.0.0.1:5127",  // The rpc host:port for the goku CLI tool
  // "rpc": { "ip": "127.0.0.1:5127", "cert": "/path/to/cert.pfx" }, // alternatively provide an object with an ip and a certificate location
  // "rpc": [ "127.0.0.1:5127" ],  // alternatively provide an array of ip's for RPC to bind to
  // "rpc": [ { "ip": "127.0.0.1:5127", "cert": "/path/to/cert.pfx" } ], // alternatively provide an array of ips with certificates
  "backend": {
    "type": "file", // the storage backend type, consul and file are currently supported
    "path": "./data/goku" // the path to storage
  },
  "hostname": "xip.io" // the root domain name used access active services running in goku
  "masterOnly": false, // if true, only accepts pushes to master branch
  "gitpath": "/tmp/path/to/repos", // the temp file path where the bare git repositories are stored
  "dockersock": "/var/run/docker.sock", // the docker daemon endpoint
  "registry": "docker.io", // the url/ip to a docker registry
  "debug": false // turns on debug mode,
}
```

### Deploying an app to Goku

1. Setup your project by adding either a `Dockerfile` or `docker-compose.yml` file in your project's root.

> This Dockerfile must expose at least one port or Goku will error out

2. Add the remote to your repo like so: `git remote add goku git@<goku server ip/hostname>:<repository name>.git`

3. Then push: `git push goku`

You will see some validation and build output as the repository is processed.

If your repository is successfully built, Goku will publish your app at `reponame.(Goku server ip).xip.io`.

### Manipulating Goku via CLI

Goku's server runs an HTTP/JSON RPC endpoint that can be manipulated with the CLI

you can set the target rpc endpoint by either setting the `GOKU_RPC_ADDRESS` environment variable or passing a `-address` flag.

By default the client runs against "127.0.0.1:5127", so for most cases no setup is necessary

`goku app list|ls`: list currently running apps

`goku app kill|rm <id>`: kill 

`goku app details <id>`: lists bound ports, the id(s) associated with the running container(s), and the repo the app belongs to

`goku app logs <id> -tail -prefix`: prints logs for an app. Optionally enable real time tailing or prefix each statement with the app name

`goku user add <name> <password>`: add a user

`goku user list`: lists the users

`goku user rm <name>` : removes a user

`goku git clear`: wipes all of the bare git repositories

`goku git ls`: lists pushed git repositories

`goku config debug <true/false>`: changes debug mode

`goku config hostname <hostname>`: changes the default host name goku uses for newly pushed services

`goku config registry <registry>`: a private docker registry path

## Contributing

Make sure you have Vagrant installed and then:

1. `vagrant up`
2. :sunglasses:

## License

MIT
