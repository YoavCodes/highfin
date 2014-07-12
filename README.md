highfin
=======
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
