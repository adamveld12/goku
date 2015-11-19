# Goku

A small, easy to install, easy to manage, scalable, get the girls, get the boys, PaaS for hobbyists.


## Features

- An admin dashboard:
  - manage and monitor deployed apps
  - manage ssh keys

- Docker compose support
    Push a repository that contains a docker-compose.yml and you're all set

- Scaling with Docker Swarm
    Cluster several Goku instances and go super sayain with docker swarm


## How to use 


### Install
Run './install.sh'. This will install Docker, pull down the latest Goku image and run it.

Goku defaults to [Docker in Docker]() mode. If you would like to run Goku such that it uses the Docker daemon
on the host, you can optionally run './install.sh host'. If you would like to run Goku under a custom domain, then use 

### Using with Git

Add the remote to your repo and simply `git push goku`. You will see some validation and build output as the repository is processed. 

If your repository is successfully built, Goku will look for a CNAME file in the root for custom domains, and also register `$REPO-$BRANCH.(Goku server ip).xip.io` as well. 

Currently Goku doesn't directly support running under a custom domain, but that will be coming.


## How does this work?

1. `git push git@server.example.com` initiates a ssh connection to the git user on the octohost.

2. The sshd server looks for the ssh public key in `/home/git/.ssh/authorized_keys` - if there's no match, it denies access. To add the public key - `cat ~/.ssh/{public_key}.pub | ssh -i ~/your-private-key.pem ubuntu@ip.address.here "sudo gitreceive upload-key {your-name-here}"`

3. It runs the command defined in `/home/git/.ssh/authorized_keys` - which is normally: `/usr/bin/gitreceive run {name} {ssh:key:finger:print}`

4. gitreceive takes over and ends up pushing the code through git archive to `/home/git/receiver`.

5. The receiver loads the main octohost configuration file by sourcing it. This configuration file is used by `/home/git/receiver` and `/usr/bin/octo`

6. The receiver takes the bare repo it gets and checks out a working repo in `/home/git/src/$NAME.git`

7. If the git branch isn't master, it will create a new container named `$NAME-$BRANCH` - if it's master, it's just named `$NAME`.

8. The receiver looks for any old containers running with the same name - it will use this information to stop it later. (We're not really using the information anymore, but it's the trigger that helps to know there's something to kill.)

9. The receiver looks for the image_id of the current running image. This will help to kill old container images.

10. If there's a Dockerfile inside `/home/git/src/$NAME.git/` then it starts to build the container. Otherwise, we're done here.

11. It looks inside the Dockerfile for an EXPOSE command - this is used to run the container and figure out how to connect to it from the outside.

12. Then the Docker container is built. If the build fails, it exits here.

13. There are a number of options that are needed to run any container, those are all gathered using the `/usr/bin/octo config:options` command detailed here.

14. Any config variables stored in Consul's KV storage are also added to the `$RUN_OPTIONS`. Config variables can be added with `/usr/bin/octo config:set $NAME/$KEY $VARIABLE`

15. Once we have all of the `$RUN_OPTIONS`, we can actually run the container.

16. Once the container is running, we look for the exposed port that's connected to the internal port.

17. If there is no exposed HTTP port, then it kills the old container and is complete.

18. This section will be removed. It was to work around a specific Docker bug that should be fixed now.

19. Any tags from the container are grabbed using `/usr/bin/octo service:tags` - and the Consul service is created with `/usr/bin/octo service:set`

20. We grab all of the possible domains and then set them in Consul with `/usr/bin/octo domains:set`.

21. If you want to launch more than 1 container when you push, you can set a key/value in Consul at `$NAME/CONTAINERS` - we check it now and launch more containers as needed.

22. Any old containers are killed.

23. nginx is reconfigured, a new `/etc/nginx/containers.conf` is written and it is reloaded.

24. If you have a private registry defined in `/etc/default/octohost`, then we push the new image there.

## Develop


## License

MIT
