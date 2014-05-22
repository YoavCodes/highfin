shark

this is the shark development environment.
to deploy all you need copy the shark binary into usr/bin
sudo /shark install should install itself onto a bare ubuntu 14.04 image
it will create a folder called /shark and config file /shark/config.json
it should always be run with sudo or as root.



install dependencies on base linux box. use nave: https://github.com/isaacs/nave


shark_setup.sh will run after vagrant installs.
the /node folder in this project contains every version of node.js binary. 	

{{create a separate vagrant job to make/install the binaries from source.. at some point}}

shark_stetup.sh will copy the version of node.js requried by shark into /usr/bin 



=== node.js binaries can be downloaded from node site

=== create npm binaries
download binaries from here
http://registry.npmjs.org/npm/-/npm-1.3.22.tgz

it's convoluted.
copy files around to achieve:
/bin/npm
/bin/npm/node_modules
/bin/npm/node_modules/npm
/bin/npm/node_modules/npm/bin   // this is actually the /bin folder again
/bin/npm/node_modules/npm/lib



=== GIT

shark can create GIT repos.
shark can manage dev keys for a repo.
When shark creates a GIT repo it copies files into the GIT hooks folder of the repos it creates. These scripts are node.js shell scrips that trigger a post request to the running shark instance 127.0.0.1:shark_port. 

hook 1. pre-commit: authenticate.
	- if shark can maintain a list of valid keys for a repo in the hooks folder itself, rather do that so shark doesn't have to respond at all to unauthenticated users.

hook 2. post-commit: build squid
	- run bamboo.js
		- runs server.js (salmon.js server) copied from server.. don't trust user's salmon.js but do respect their version dependency
		- runs tests.js. if failure then communicate that to tiger somehow or just exit or something.
	- tiger will then begin gracefully redeploying the squid.. which basically involves shutting down the old squid



every useraccount/project ie: chalkhq_highfin has a linux user and password.
the password is a hash salted with the name of the useraccount_project.
in the ~/.ssh/authorized_keys is where dev keys are managed for that repo. shark can simply add/remove users' access to projects this way

on osx, or in vagrant vm, in your project folder make a dev_key folder.
in terminal run 
ssh-keygen -C "yoav@chalkhq.com"
and drag the dev_key folder in from finder. set a passphrase if you want.
node salmon push can be run from inside the vm, should use config for account/project name to do the push.. in fact, when salmond.js vagrant up is run it should generate the key for you based on the account/project name


==============left off:===========
- git server working.
- move git user creation, repo creation to shark.js node app function
- add RSA key generation to salmon.js vagrant setup and modify folder structure so keys are not in the git repo
- add salmon.js push command to perform a git commit to the running shark server (use osx hosts file to redirect git.highf.in to  10.10.10.2)


-- ssh works with specific keyfile
-- git (tower and terminal) do not.
-- git can NOT use specific key file without altering ssh configuration on the client.
-- since we only want to git push to dev as a salmon command, we can set the global dev key inside the vm just fine
-- until then either use passwords.. OR even better, get working on the salmonjs vagrant up using the local shark ip and a known shared dev_key.




ssh -i ./dev_key/key test8@10.10.10.2 # from the project folder

ssh -i /vagrant/dev_key/key test8@10.10.10.2






git_chalkhq_highfin@10.10.10.2:/home/git_chalkhq_highfin/dev-next.git

in your git client works

salmon push next|current
	1. run unit tests locally
	2. do a git push using key to your project folder

separate bash scripts in vagrant bootstrap into files in /usr/bin and trigger them that way, so that if a user just wants to regenerate their dev_key, or update node.js or something they can just type 
salmon use node 1.0.10 
	-- change to use downloaded node.js, handle folder reshuffle in script so script can add node versions seemlessly
salmon generate dev_key
	-- message to register public key with shark
salmon push current
salmon push next
salmon deploy current qa
salmon deploy next qa
salmon deploy current prod
salmon deploy next prod
salmon maintenance next/current dev/qa / prod