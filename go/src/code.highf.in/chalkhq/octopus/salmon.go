package main

import (
	"os/exec"
)

func UpdateSalmon() {
	_ = exec.Command("rm", "-r", "/octopus/salmon").Run()

	// it needs to do a shallow clone of salmon to /octopus/salmon,
	_ = exec.Command("git", "clone", "--depth", "1", "https://github.com/YoavGivati/salmon", "/octopus/salmon/").Run()
	// delete the .git folder, run git init, git add ., git commit -m "new salmon repo"
	_ = exec.Command("rm", "-R", "/octopus/salmon/.git").Run()

	cmd := exec.Command("git", "init")
	cmd.Dir = "/octopus/salmon"
	_ = cmd.Run()

	cmd = exec.Command("git", "add", ".")
	cmd.Dir = "/octopus/salmon"
	_ = cmd.Run()

	cmd = exec.Command("git", "commit", "-m", `"new salmon repo"`)
	cmd.Dir = "/octopus/salmon"
	_ = cmd.Run()

	cmd = exec.Command("git", "branch", "dev-next")
	cmd.Dir = "/octopus/salmon"
	_ = cmd.Run()

	cmd = exec.Command("git", "checkout", "dev-next")
	cmd.Dir = "/octopus/salmon"
	_ = cmd.Run()

	cmd = exec.Command("git", "branch", "-D", "master")
	cmd.Dir = "/octopus/salmon"
	_ = cmd.Run()

	cmd = exec.Command("git", "sybolic-ref", "HEAD", "refs/heads/dev-next")
	cmd.Dir = "/octopus/salmon"
	_ = cmd.Run()

	// todo: use link instead to make updating post-receive easier and save hd
	_ = exec.Command("cp", "-f", "/vagrant/go/bin/post-receive", "/octopus/salmon/.git/hooks/post-receive").Run()

	// todo: only for testing, remove this line. salmon repo should include this in the hooks folder
	_ = exec.Command("cp", "-f", "/vagrant/go/bin/post-receive", "/coral/chalkhq/nodetest/code.git/hooks/post-receive").Run()

	//sed -i 's/bare = false/bare = true/' /coral/chalkhq/nodetest/code.git/config
	_ = exec.Command("sed", "-i", "s/bare = false/bare = true/", "/octopus/salmon/.git/config").Run()

	// todo: this would be triggered by a git postreceive hook in the salmon repo
	// should also create devnext and devcurrent branches
	// then git create() will simply copy this new salmon repo's .git folder to a user's project git folder code.git
	// git symbolic-ref HEAD refs/heads/dev-next make the dev-next branch default, git branch -D master
	/*
		git branch dev-next
		git branch -D master
		git symbolic-ref HEAD refs/heads/dev-next
		// other branches are delete and created during deploys to various environments
	*/
}
