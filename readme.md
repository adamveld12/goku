# Goku

A small, easy to install, easy to manage PaaS for hobbyists.


## How to use 

Install

* Nginx
* Docker

`go get` and `go install` goku, then run `goku -ssh ":2222"`


### Using with Git

Add the remote to your repo and simply `git push goku`. You will see some validation and build output as the repository is processed. 

If your repository is successfully built, Goku will publish your app at `reponame.(Goku server ip).xip.io`.


## License

MIT
