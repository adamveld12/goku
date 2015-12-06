# -*- mode: ruby -*-
# vi: set ft=ruby :

# All Vagrant configuration is done below. The "2" in Vagrant.configure
# configures the configuration version (we support older styles for
# backwards compatibility). Please don't change it unless you know what
# you're doing.
Vagrant.configure(2) do |config|
  # The most common configuration options are documented and commented below.
  # For a complete reference, please see the online documentation at
  # https://docs.vagrantup.com.

  config.vm.box = "ubuntu/trusty64"

  # Create a forwarded port mapping which allows access to a specific port
  # within the machine from a port on the host machine. In the example below,
  # accessing "localhost:8080" will access port 80 on the guest machine.
  
  # # SSH
  config.vm.network "forwarded_port", guest: 2222, host: 2223
  # # NGINX
  config.vm.network "forwarded_port", guest: 80, host: 3000

  # Create a private network, which allows host-only access to the machine
  # using a specific IP.
  config.vm.network "public_network", ip: "192.168.99.100"

  # Create a public network, which generally matched to bridged network.
  # Bridged networks make the machine appear as another physical device on
  # your network.
  #config.vm.network "public_network"

  # Share an additional folder to the guest VM. The first argument is
  # the path on the host to the actual folder. The second argument is
  # the path on the guest to mount the folder. And the optional third
  # argument is a set of non-required options.
  config.vm.synced_folder ".", "/go/src/github.com/adamveld12/goku"

  # Provider-specific configuration so you can fine-tune various
  # backing providers for Vagrant. These expose provider-specific options.
  # Example for VirtualBox:
   config.vm.provider "virtualbox" do |vb|
     vb.memory = "1024"
   end

  # Enable provisioning with a shell script. Additional provisioners such as
  # Puppet, Chef, Ansible, Salt, and Docker are also available. Please see the
  # documentation for more information about their specific syntax and use.
   config.vm.provision "shell", inline: <<-SHELL
    sudo apt-key adv --keyserver hkp://p80.pool.sks-keyservers.net:80 --recv-keys 58118E89F3A912897C070ADBF76221572C52609D
     sudo echo "deb https://apt.dockerproject.org/repo ubuntu-trusty main" > /etc/apt/sources.list.d/docker.list
     sudo apt-get update && sudo apt-get purge lxc-docker && sudo apt-cache policy docker-engine
     sudo apt-get install -y nginx \
                             git \
                             linux-image-generic-lts-trusty \
                             linux-image-extra-$(uname -r) \
                             docker-engine
     rm -rf /vagrant_data
     cd /usr/local/bin/
     curl -s https://storage.googleapis.com/golang/go1.5.1.linux-amd64.tar.gz | tar zx
     sudo cat > /home/vagrant/.bashrc <<RC
     export GOPATH=/go
     export GOROOT=/usr/local/bin/go
     export PATH=$PATH:/usr/local/bin/go/bin
RC
     sudo usermod -aG docker vagrant
     sudo chown -R vagrant /go
     sudo chown -R vagrant /etc/nginx

   SHELL
end
