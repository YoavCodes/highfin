package main

import (
	"fmt"
	"os"
	"os/exec"
)

func CreateStartupScript() {
	startup_path := "/etc/init.d"
	_ = exec.Command("mkdir", "-p", startup_path).Run()
	file_path := startup_path + "/fishtank"

	file, err := os.Create(file_path)
	defer file.Close()

	if err != nil {
		fmt.Println("error: " + err.Error())
		return
	}

	contents := `` +
		`#! /bin/sh
### BEGIN INIT INFO
# Provides:          fishtank
# Required-Start:    $remote_fs $syslog
# Required-Stop:     $remote_fs $syslog
# Default-Start:     2 3 4 5
# Default-Stop:      0 1 6
# Short-Description: http://highf.in/fishtank
# Description:       This file should be used to construct scripts to be
#                    placed in /etc/init.d.
### END INIT INFO

# Author: Yoav Givati <dev@yoavgivati.com>
#
# Please remove the "Author" lines above and replace them
# with your own name if you copy and modify this script.

# Do NOT "set -e"

# PATH should only include /usr/* if it runs after the mountnfs.sh script
PATH=/sbin:/usr/sbin:/bin:/usr/bin
DESC="fishtank"
NAME=fishtank
DAEMON=/usr/bin/$NAME
DAEMON_ARGS=""
PIDFILE=/var/run/$NAME.pid
SCRIPTNAME=/etc/init.d/$NAME

# Exit if the package is not installed
[ -x "$DAEMON" ] || exit 0

# Read configuration variable file if it is present
[ -r /etc/default/$NAME ] && . /etc/default/$NAME

# Load the VERBOSE setting and other rcS variables
. /lib/init/vars.sh

# Define LSB log_* functions.
# Depend on lsb-base (>= 3.2-14) to ensure that this file is present
# and status_of_proc is working.
. /lib/lsb/init-functions

#
# Function that starts the daemon/service
#
do_start()
{
	# Return
	#   0 if daemon has been started
	#   1 if daemon was already running
	#   2 if daemon could not be started
	start-stop-daemon --start --quiet --pidfile $PIDFILE --exec $DAEMON --test > /dev/null \
		|| return 1
	start-stop-daemon --start --quiet --pidfile $PIDFILE --exec $DAEMON -- \
		$DAEMON_ARGS \
		|| return 2
	# Add code here, if necessary, that waits for the process to be ready
	# to handle requests from services started subsequently which depend
	# on this one.  As a last resort, sleep for some time.
}

#
# Function that stops the daemon/service
#
do_stop()
{
	# Return
	#   0 if daemon has been stopped
	#   1 if daemon was already stopped
	#   2 if daemon could not be stopped
	#   other if a failure occurred
	start-stop-daemon --stop --quiet --retry=TERM/30/KILL/5 --pidfile $PIDFILE --name $NAME
	RETVAL="$?"
	[ "$RETVAL" = 2 ] && return 2
	# Wait for children to finish too if this is a daemon that forks
	# and if the daemon is only ever run from this initscript.
	# If the above conditions are not satisfied then add some other code
	# that waits for the process to drop all resources that could be
	# needed by services started subsequently.  A last resort is to
	# sleep for some time.
	start-stop-daemon --stop --quiet --oknodo --retry=0/30/KILL/5 --exec $DAEMON
	[ "$?" = 2 ] && return 2
	# Many daemons don't delete their pidfiles when they exit.
	rm -f $PIDFILE
	return "$RETVAL"
}

#
# Function that sends a SIGHUP to the daemon/service
#
do_reload() {
	#
	# If the daemon can reload its configuration without
	# restarting (for example, when it is sent a SIGHUP),
	# then implement that here.
	#
	start-stop-daemon --stop --signal 1 --quiet --pidfile $PIDFILE --name $NAME
	return 0
}

case "$1" in
  start)
	[ "$VERBOSE" != no ] && log_daemon_msg "Starting $DESC" "$NAME"
	do_start
	case "$?" in
		0|1) [ "$VERBOSE" != no ] && log_end_msg 0 ;;
		2) [ "$VERBOSE" != no ] && log_end_msg 1 ;;
	esac
	;;
  stop)
	[ "$VERBOSE" != no ] && log_daemon_msg "Stopping $DESC" "$NAME"
	do_stop
	case "$?" in
		0|1) [ "$VERBOSE" != no ] && log_end_msg 0 ;;
		2) [ "$VERBOSE" != no ] && log_end_msg 1 ;;
	esac
	;;
  status)
	status_of_proc "$DAEMON" "$NAME" && exit 0 || exit $?
	;;
  #reload|force-reload)
	#
	# If do_reload() is not implemented then leave this commented out
	# and leave 'force-reload' as an alias for 'restart'.
	#
	#log_daemon_msg "Reloading $DESC" "$NAME"
	#do_reload
	#log_end_msg $?
	#;;
  restart|force-reload)
	#
	# If the "reload" option is implemented then remove the
	# 'force-reload' alias
	#
	log_daemon_msg "Restarting $DESC" "$NAME"
	do_stop
	case "$?" in
	  0|1)
		do_start
		case "$?" in
			0) log_end_msg 0 ;;
			1) log_end_msg 1 ;; # Old process is still running
			*) log_end_msg 1 ;; # Failed to start
		esac
		;;
	  *)
		# Failed to stop
		log_end_msg 1
		;;
	esac
	;;
  *)
	#echo "Usage: $SCRIPTNAME {start|stop|restart|reload|force-reload}" >&2
	echo "Usage: $SCRIPTNAME {start|stop|status|restart|force-reload}" >&2
	exit 3
	;;
esac

exit 0
`

	file.WriteString(contents + "\n\n")
	fmt.Println("/etc/init.d/fishtank daemon script created")
}

func CreateDefaultRevision(project string, branch string) {
	default_path := "/srv/www/" + project + "/default"
	_ = exec.Command("mkdir", "-p", default_path).Run()
	file := default_path + "/index.html"

	conf, err := os.Create(file)
	defer conf.Close()

	if err != nil {
		fmt.Println("error: " + err.Error())
		return
	}

	contents := `` +
		`<!DOCTYPE html>
<html>
<head>
<meta http-equiv="Content-Type" content="text/html; charset=UTF-8" />
<title>` + project + `</title>
</head>

<style type="text/css">
*, html, body {
	background-color: #222;	
}

#message {
	width: 800px;
	margin: 40px 0 0 140px;;	
	font-family:Helvetica, Arial, sans-serif;
	font-size: 60px;
	font-weight: bold;
	color: #fff;
}

.grey {
	color: #666;	
	font-size: 12px;
}
</style>

<body>

<div id="message">
<p>` + project + ` <span class="grey">Welcome to Fishtank, git push your ` + branch + ` branch to update it</span></p>
</div>


</body>
</html>
`

	conf.WriteString(contents)
	fmt.Println("Default site created")
}

func CreateNginxSiteConf(project string) {
	file := "/etc/nginx/sites-enabled/" + project + ".conf"

	conf, err := os.Create(file)
	defer conf.Close()

	if err != nil {
		fmt.Println("error: " + err.Error())
		return
	}

	contents := `` +
		`server { 
	listen 80; 
	server_name ` + project + `; 
	root /srv/www/` + project + `/default/;  # default will contain a hello world .html
	index index.html; 

	location = /favicon.ico { 
		log_not_found off; 
		access_log off; 
	} 

	location = /robots.txt { 
		allow all; 
		log_not_found off; 
		access_log off; 
	} 

	location / { 
		add_header X-Frame-Options SAMEORIGIN; 
		#first try a file, then a directory, then an index file 
		try_files $uri $uri/ /index.html; 
	} 

	#nodejs api 
	location /0 { 
		proxy_set_header X-Real-IP $remote_addr; 
		proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for; 
		proxy_set_header Host $http_host; 
		proxy_set_header X-NginX-Proxy true; 
		proxy_pass http://127.0.0.1:8000/; 
		proxy_redirect off; 
	} 

	location ~* \\.(js|css|png|jpg|jpeg|gif|ico)$ { 
		expires max; 
		log_not_found off; 
	} 
}
`

	conf.WriteString(contents)
	fmt.Println("nginx site.conf created")
}

// store it in the projects struct
func CreatePostRecieveHook(project string, branch string) {
	file_path := "/srv/coral/" + project + "/code.git/hooks/post-receive"
	file, err := os.Create(file_path)
	defer file.Close()

	if err != nil {
		fmt.Println("error: " + err.Error())
		return
	}

	contents := `` +
		`#!/bin/bash ` + "\n" +
		`while read oldrev newrev refname ` + "\n" +
		`do ` + "\n" +
		`    branch=$(git rev-parse --symbolic --abbrev-ref $refname) ` + "\n" +
		`    if [ "` + branch + `" == "$branch" ]; then ` + "\n" +
		`        # Do something ` + "\n" +
		`        curl -s "http://127.0.0.1:` + config.Port + `/post-receive?project=` + project + `&old_rev=$oldrev&new_rev=$newrev&branch=$branch"` + "\n" +
		`    fi ` + "\n" +
		`done ` + "\n"

	file.WriteString(contents + "\n\n")
	fmt.Println("post-receive created")
}
