#!/bin/bash
salmonNote() {
	echo -e "\033[0;97;41m ===SALMON=== \033[0m\033[0;97;46m $1 \033[0m" # ===white on red=== white on cyan
}


PROJECT_DIR=/vagrant
USER_HOME=/home/vagrant

# configuration, should load from external files
HIGHFIN_IP="highf.in"
COMPANY=""
PROJECT=""
DEV_KEY_VALID=false
DEV_VALID_RESPONSE=""
EMAIL=""

if [ ! $1 ];
then
	salmonNote "Please enter a command, valid commands are one of [clone]"
	exit 0
fi

case "$1" in

	"clone-docs")
		salmonNote "removing old /salmon folder contents"
		sudo rm -f -r /salmon/*

		for BRANCH in "docs" "examples/simple-site"
		do			
			# BRANCH corresponds folder hierarchy	
			# github puts branch examples/simple-site in the folder salmon-examples-simple-site in the zip file
			BRANCH_FOLDER=$(echo "salmon-$BRANCH" | sed -r 's/[\/]/-/g')
			
			salmonNote "fetching $BRANCH"
			sudo mkdir -p /salmon/tmp
			cd /salmon/tmp
			sudo wget -qO- -O tmp.zip https://github.com/YoavGivati/salmon/archive/$BRANCH.zip && sudo unzip tmp.zip &> /dev/null  && sudo rm tmp.zip
			sudo mkdir -p /salmon/$BRANCH
			sudo mv $BRANCH_FOLDER/* /salmon/$BRANCH
			sudo rm -r $BRANCH_FOLDER
			cd /salmon/$BRANCH/salmon
			sudo npm install
		done
		

	;;

	"clone")
		
		salmonNote "Setting up project"

		COMPANY=$2
		PROJECT=$3
		EMAIL=$4

		while [[ ! $COMPANY =~ ^[a-z]+$ ]]; do
			echo "Please enter company name:"
			read COMPANY
		done

		while [[ ! $PROJECT =~ ^[a-z]+$ ]]; do
			echo "Please enter project name:"
			read PROJECT
		done

		while [[ ! $EMAIL =~ ^.+$ ]]; do
			echo "Please enter your email address:"
			read EMAIL
		done

		salmonNote "Configure dev key"
		# generate a dev key, if one doesn't already exist in the folder. you should also git ignore the folder's contents.
		if [ ! -e /vagrant/dev_key/key ] && [ ! -e /vagrant/dev_key/key.pub ];
		then	
			salmonNote 'Generating new dev key'
			#"/vagrant/dev_key/key" | ssh-keygen -C "$EMAIL"
			#"/vagrant/dev_key/key" | ssh-keygen -C "yoav@chalkhq.com"
			mkdir /vagrant/dev_key
			ssh-keygen -t rsa -N "" -f /vagrant/dev_key/key -C "$EMAIL"
		else
			salmonNote  'Sweet, dev key already exists'
		fi

		KEY_PUB=$(cat /vagrant/dev_key/key.pub)

		while [[ $DEV_KEY_VALID == false ]]; do
			salmonNote "Verifying dev has access to project.."				

			DEV_VALID_RESPONSE=$(curl -sL -w "%{http_code}\\n" --data "company=$COMPANY&project=$PROJECT&email=$EMAIL&key=$KEY_PUB" "http://highf.in/0" -o /dev/null) 
			
			if [ ${DEV_VALID_RESPONSE} == 200 ] ; then
				salmonNote "Our records show that you have access to this project"
				DEV_KEY_VALID=true
			else
				salmonNote "==== Bloop bloop bloop"
				salmonNote "Our records show that you do not have access to this project"				
				salmonNote "Please register your dev key with ${COMPANY}'s ${PROJECT} project"
				salmonNote "You can do this by signing into http://highf.in" 
				salmonNote "or contacting the admin of your company."
				echo -e "============\\n\\n"
				echo $KEY_PUB
				echo -e "\\n\\n============"
				read -p "Press enter to verify dev access"
			fi			
		done

		#sudo mkdir -p /home/vagrant/.ssh && sudo touch /home/vagrant/.ssh/authorized_keys
		#sudo sh -c "cat /vagrant/dev_key/key.pub >> /home/vagrant/.ssh/authorized_keys"

		salmonNote 'Expose dev key to GIT'
		sudo su - root /bin/bash -c 'sudo echo "IdentityFile /vagrant/dev_key/key" >> /etc/ssh/ssh_config'


		salmonNote "Cloning ${PROJECT} project from ${COMPANY}"
		sudo rm -R /vagrant/project
		### note: connecting to local shark requires host key verification. self-signed cert. isn't an issue in production, so just accept if local
		# git clone git_chalkhq_highfin@10.10.10.2:project.git /vagrant/project
		# it should prompt yes/no, sometimes it fails to do that when this command is executed in this script
		salmonNote 'cloneing project repo'
		salmonNote "git clone git_${COMPANY}_${PROJECT}@${HIGHFIN_IP}:project.git /vagrant/project"
		sudo git clone git_${COMPANY}_${PROJECT}@${HIGHFIN_IP}:project.git /vagrant/project		

		salmonNote "Installing node.js packages.."
		# Sudo su to get a fresh environment when running command
		sudo su - vagrant /bin/bash -c "cd ${PROJECT_DIR}/project/server; sudo npm install;"


		##
		#	Populate mongodb
		##
		##todo: move mongo fixtures to project/server/fixtures/mongo
		# if [ -d /vagrant/project/conf/local/mongodb ];
		# then
		# 	echo "restoring mongo"
		# 	for i in /vagrant/project/conf/local/mongodb/*; do
		# 		p=`expr $i : '\/vagrant\/project\/conf/\local\/mongodb\/\([^\/]*\)'`
		# 		echo "restoring "$p
		# 		mongorestore -d $p /vagrant/project/conf/local/mongodb/$p
		# 	done
		# fi

		sudo service nginx restart	
		salmonNote "vagrant up completed successfully"

		



		

		

		exit 0
	;;

esac


exit 0