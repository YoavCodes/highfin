highfin
=======

The purpose of this project is to build a highly scalable flexible deployment infrastructure. currently supports node.js, golang, and mongodb. I'm leaving this as a functional prototype so there are lots of optimizations to be made with regard to streaming data, proper error handling, unit tests, etc.

I'm adding the code in this repo to the public domain as is without warranty. 

This project is a work in progress. The next thing to do is to make sure Octopus doesn't redeploy a mongo instance unless the ram/mongo version/cpu or other specs change and copy the docker attached data volume to the new shark when that happens. -.json config file should support specifying ram/cpu shares etc. Sharks should let Octoups know about their available resources so Octopus can select an available shark when deploying/redeploying. jellyfish should reject connections from other users' deploys. and find a way to set permissions on attached volumes so that only a given user/project can access a given volume.


Open 6 terminal windows

t0:

$vagrant up

$vagrant ssh guppy

t1:

$vagrant ssh guppy

$guppy run guppy

t2:

$vagrant ssh octopus

$guppy run octopus

t3:

$vagrant ssh shark0

$guppy run shark

t4:

$vagrant ssh shark1

$guppy run shark

t5:

$vagrant ssh squid

$guppy run squid


t0:

// this will create a new git repo on octopus

$guppy create

// this will pull that repo into your guppy vm /vagrant/code

$guppy get

// this will tell octopus to deploy the latest code from the dev-next branch

$guppy push dev-next

// configure your /etc/hosts file

10.10.10.5      app1.test

10.10.10.5      app2.test

// you should be able to hit app1.test or app2.test in your browser once the deploy is complete

