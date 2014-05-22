#!/bin/bash
salmonNote() {
	echo -e "\033[0;97;41m ===SALMON=== \033[0m\033[0;97;46m $1 \033[0m" # ===white on red=== white on cyan
}

if [ ! -e /home/vagrant/setupcomplete ]; then

###
# versioning  x.yz = x:release version.y:dev version z:testable change
###
SALMON_BASEBOX_VERSION=0.23

GO_VERSION=1.2.1
NODE_VERSION=0.10.24
NPM_VERSION=1.3.22

salmonNote "Salmon Basebox v$SALMON_BASEBOX_VERSION"
salmonNote "Environment installation is beginning. This may take a few minutes.."

#todo: note, this causes sudo: unable to resolve host salmonbox /etc/hosts.. maybe other stuff
#sudo sh -c "echo 'salmonbox' > /etc/hostname"

##
#	Install core components
##

salmonNote "Updating package repositories.."
#apt-get update

salmonNote "Installing required packages.."
#apt-get -y install git libpq-dev pkg-config make cmake flex build-essential g++ nfs-common mongodb-10gen htop iftop curl unzip
#apt-get -y install git pkg-config curl unzip
apt-get -y install git

salmonNote "Install Golang"
if [ ! -e /vagrant/lib/go${GO_VERSION}.linux-amd64.tar.gz  ]; then
	salmonNote "downloading go 1.2.1 binaries from Google"
	wget -P /vagrant/lib/ https://go.googlecode.com/files/go${GO_VERSION}.linux-amd64.tar.gz 
	# wget -P /vagrant/lib/ https://go.googlecode.com/files/go1.2.1.linux-amd64.tar.gz 
fi

tar -C /usr/local -xzf /vagrant/lib/go${GO_VERSION}.linux-amd64.tar.gz 
#tar -C /usr/local -xzf /vagrant/lib/go1.2.1.linux-amd64.tar.gz 


salmonNote 'Install node & npm binaries'
sudo cp /vagrant/lib/node-$NODE_VERSION /usr/bin/node
#curl https://npmjs.org/install.sh | sudo sh
 
sudo mkdir /usr/bin/node_modules
 
sudo cp /vagrant/lib/npm-$NPM_VERSION/npm /usr/bin/npm
sudo cp -R /vagrant/lib/npm-$NPM_VERSION/node_modules/* /usr/bin/node_modules

# export the go and workspaces bin folders to PATH
#sudo su - root /bin/bash -c 'echo PATH="/vagrant/go/bin:/usr/local/go/bin:/guppy:$PATH\"" > /etc/environment'
echo 'PATH="/vagrant/shark/go/bin:/vagrant/salmon/go/bin:/vagrant/guppy/go/bin:/vagrant/code/go/bin:/usr/local/go/bin:/guppy:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/usr/games:/usr/local/games"' > /etc/environment

# will look in your vagrant bin, then go's bin, then guppy, then everywhere else.
# sudo will look everywhere else first

# setup your workspace
#sudo su - root /bin/bash -c 'echo "GOPATH=\"/vagrant/go\"" >> /etc/environment'
echo 'GOPATH="/vagrant/shark/go:/vagrant/salmon/go:/vagrant/guppy/go:/vagrant/go:/vagrant/code"' >> /etc/environment

# bind port 80 to 8080, so apps can listen on 80 via 8080 ie: without sudo
iptables -t nat -A PREROUTING -p tcp --dport 80 -j REDIRECT --to-port 8080

# setup guppy storage folder
# need admin permissions for this so we do it here and not in guppy itself
mkdir /guppy
chown -R vagrant:admin /guppy

# give guppy ownership of "binary" file, so it can replace itself ie: guppy finish
touch /usr/bin/guppy 
chown -R vagrant:admin /usr/bin/guppy

#get salmon so we can work on it
guppy get-salmon

# finish the basebox ie: prep it for .box creation in case we don't even want to ssh in
guppy finish

# add guppy ssh key
sudo su - root /bin/bash -c 'sudo echo "IdentityFile /vagrant/keys/guppy" >> /etc/ssh/ssh_config'

#guppy get highf.in examples

#salmonNote "Installing latest Salmon docs & demos"

#sudo salmon clone-docs

#salmonNote "Installing node.js packages.."

salmonNote "Basebox is installed, run: vagrant package --output fry-basebox_v${SALMON_BASEBOX_VERSION}.box"

exit 0

fi

salmonNote "Basebox is already setup, run destroy/up to start fresh"

exit 0