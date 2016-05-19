# -*- mode: ruby -*-
# vi: set ft=ruby :

Vagrant.configure(2) do |config|
  config.vm.box = "ubuntu/trusty64"

  # SSH
  config.vm.network "forwarded_port", guest: 2222, host: 2223
  # NGINX
  config.vm.network "forwarded_port", guest: 80, host: 3000

  # Create a private network, which allows host-only access to the machine
  # using a specific IP.
  config.vm.network "public_network", ip: "192.168.99.100"

  # Create a public network, which generally matched to bridged network.
  # Bridged networks make the machine appear as another physical device on
  # your network.
  #config.vm.network "public_network"

  config.vm.synced_folder "", "/go/src/github.com/adamveld12/goku"

  config.vm.provider "virtualbox" do |vb|
    vb.memory = "1024"
  end

  config.vm.provision "shell", inline: <<-SHELL
     sudo apt-key adv --keyserver hkp://p80.pool.sks-keyservers.net:80 --recv-keys 58118E89F3A912897C070ADBF76221572C52609D;
     sudo echo "deb https://apt.dockerproject.org/repo ubuntu-trusty main" > /etc/apt/sources.list.d/docker.list;
     sudo apt-get update \
     && sudo apt-get purge lxc-docker \
     && sudo apt-cache policy docker-engine;

     sudo apt-get install -y nginx \
                             git \
                             linux-image-generic-lts-trusty \
                             linux-image-extra-$(uname -r) \
                             docker-engine;
     rm -rf /vagrant_data;

    curl -L https://github.com/docker/compose/releases/download/1.6.2/docker-compose-`uname -s`-`uname -m` > /usr/local/bin/docker-compose;
    chmod +x /usr/local/bin/docker-compose;

     cd /usr/local/bin/;
     curl -s https://storage.googleapis.com/golang/go1.5.1.linux-amd64.tar.gz | tar zx;
     sudo cat > /home/vagrant/.bashrc <<RC
     export GOPATH=/go
     export GOROOT=/usr/local/bin/go
     export PATH=$PATH:/usr/local/bin/go/bin
RC

     sudo usermod -aG docker vagrant;
     sudo chown -R vagrant /go;
     sudo chown -R vagrant /etc/nginx;
SHELL
end
