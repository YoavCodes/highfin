#!/bin/bash
# if stdin is a terminal
if ! [ -t 0 ]; then
  read -a ref # read oldrev newrev and full/branch/name into ref array
fi
# split full/branch/name by "/" into REF array
IFS='/' read -ra REF <<< "${ref[2]}"

oldrev="${ref[0]}"
newrev="${ref[1]}"
branch="${REF[2]}" # get branch name

# set project pathname
project_name="highf.in" # set this when this file is created

# set project paths
project_folder="/srv/production/$project_name/"
new_rev_path="$project_folder/$newrev"
old_rev_path="$project_folder/$oldrev"

nginx_conf_file="/srv/nginx/$project_name.conf"
log_file="/srv/logs/production/$project_name.log"

app_config_file="$new_rev_path/application/config.js"

###############################################
## switch to new code and respawn everything ##

# ensure user pushed to production branch
if [ "production" == "$branch" ]; then	

	############
	echo "checking out latest revision..."
	# create directory for new rev
	mkdir -p "$new_rev_path"
	# checkout branch into project folder
  	git --work-tree=$new_rev_path checkout production -f

  	############
  	echo "configuring nginx..."
    #touch "$nginx_conf_file"
    # find next available port
    for port in $(seq 8000 65000); 
    do 
      echo "trying port $port";
      echo -ne "\035" | telnet 127.0.0.1 $port > /dev/null 2>&1; 
      [ $? -eq 1 ] && echo "unused $port" && port=$port break; 
    done
  	# change root (public) directory to new revision folder
  	sed -i.bak "s:root.*;:root /srv/production/$project_name/$newrev/public/;:" "$nginx_conf_file";
    # change nginx port redirection
  	sed -i.bak "s+proxy_pass http://.*;+proxy_pass http://127.0.0.1:$port;+" "$nginx_conf_file";
    # sed -i.bak "s+proxy_pass http://.*;+proxy_pass http://127.0.0.1:8001;+" "/srv/nginx/highfin.me.conf"
    #change app port number
    sed -i.bak "s/port:.*,/port: $port,/" "$app_config_file"


  	############
  	echo "running node"
  	# run new node
  	touch "$log_file"
  	# do this in NODE
  	sudo su - highfin /bin/bash -c "forever $new_rev_path/server/highfin_server.js -l $log_file > /dev/null 2>&1 &"
    # sudo su - highfin /bin/bash -c "forever /srv/production/highfin.me/9504987ed4e40949c80c4c2e9222b8fb3d1d5e1f/node/high.js -l /srv/logs/www/highfin.me.log > /dev/null 2>&1"

  	# hot reload nginx configuration
  	sudo nginx -s reload
  	

  	############
  	echo "removing old revision"
  	rm -R "$old_rev_path"

  	echo "production successfully deployed!"
fi