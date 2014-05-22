clone salmon into /vagrant/salmon from Github to work on it manually after cloning guppy
The repos are not technically linked, but they need to be worked on in the same vm, since the vm needs the dependencies and environment to run both of them. ie: Guppy bootstraps the fry-box, then manages all operations in the fry-box including running salmon apps.
/vagrant/code is for testing cloneing a user project from shark


/vagrant/go: guppy development, and fry-box dependencies / tools like Godep

/vagrant/salmon: actually a clone of the salmon repo

/vagrant/code: a clone of a user's project derived from salmon by shark