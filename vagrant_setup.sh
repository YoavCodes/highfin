GO_VERSION=1.4
NODE_VERSION=0.10.24
NPM_VERSION=1.3.22
# get this from prompt in vagrant up
EMAIL="git@highf.in"

note() {
	echo -e "\033[0;97;41m ===vagrant=== \033[0m\033[0;97;46m $1 \033[0m" # ===white on red=== white on cyan
}

## todo all of this, including installing golang should be migrated to shark bootstrap

note 'Register apt repos [docker]'
sudo sh -c "wget -qO- https://get.docker.io/gpg | apt-key add -"
sudo sh -c "echo deb http://get.docker.io/ubuntu docker main > /etc/apt/sources.list.d/docker.list"

note 'Update apt-get'
sudo apt-get update
sudo apt-get upgrade

note 'Install dependencies [docker, curl, git]'
sudo apt-get -y install lxc-docker curl git

#make sure all ssh host keys are generated a single missing key will prevent ssh access
ssh-keygen -A

if [ ! -e /vagrant/lib/go${GO_VERSION}.linux-amd64.tar.gz  ]; then
	note "downloading go 1.4 binaries from Google"
	wget -P /vagrant/lib/ https://storage.googleapis.com/golang/go${GO_VERSION}.linux-amd64.tar.gz	
	# wget -P /vagrant/lib/ https://storage.googleapis.com/golang/go1.4.linux-amd64.tar.gz
fi

tar -C /usr/local -xzf /vagrant/lib/go${GO_VERSION}.linux-amd64.tar.gz 


#note: guppy should install and run whichever version is stipulated in -.json file
#		it should install the latest version as main
# note 'Install node & npm binaries'
# sudo cp /vagrant/lib/node-$NODE_VERSION /usr/bin/node
 
# sudo mkdir /usr/bin/node_modules
 
# sudo cp -Rf /vagrant/lib/npm-$NPM_VERSION/npm /usr/bin/npm
# sudo cp -Rf /vagrant/lib/npm-$NPM_VERSION/node_modules/* /usr/bin/node_modules

# export the go and workspaces bin folders to PATH
#sudo su - root /bin/bash -c 'echo PATH="/vagrant/go/bin:/usr/local/go/bin:/guppy:$PATH\"" > /etc/environment'
P='PATH=/vagrant/go/bin:/usr/local/go/bin:/guppy:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/usr/games:/usr/local/games'
echo $P >> /home/vagrant/.profile
echo $P >> /root/.profile
echo $P >> /etc/environment

# will look in your vagrant bin, then go's bin, then guppy, then everywhere else.
# sudo will look everywhere else first

# setup your workspace
#sudo su - root /bin/bash -c 'echo "GOPATH=\"/vagrant/go\"" >> /etc/environment'
G='GOPATH=/vagrant/go'
echo $G >> /home/vagrant/.profile
echo $G >> /root/.profile
echo $G >> /etc/environment


note 'set global git user'
git config --global user.email "git@highf.in"
git config --global user.name "HighFin GIT Fish"

# make user switch to root and cd to /vagrant 
echo "cd /vagrant;" >> /home/vagrant/.profile
echo "sudo su;" >> /home/vagrant/.profile



# bind port 80 to 8080, so apps can listen on 80 via 8080 ie: without sudo
#iptables -t nat -A PREROUTING -p tcp --dport 80 -j REDIRECT --to-port 8080
#iptables-save > /var/iprules.fw
#echo "/sbin/iptables-restore < /var/iprules.fw" > /etc/rc.local
#echo "exit 0" >> /etc/rc.local