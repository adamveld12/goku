# Goku

A small, easy to install, easy to manage PaaS for hobbyists.

## Contributing

1. `docker-machine start default && eval $(docker-machine env default)`
2. `make up`
3. :sunglasses:

## How to use 

Install

* Nginx
* Docker

`go get` and `go install` goku, then run `goku -ssh ":2222"`


### Using with Git

Add the remote to your repo 

> git remote add goku git@<repo ip/hostname>:<repository name>.git


Then push
> git push goku

You will see some validation and build output as the repository is processed.

If your repository is successfully built, Goku will publish your app at `reponame.(Goku server ip).xip.io`.


## License

MIT
