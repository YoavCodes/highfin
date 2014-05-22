package util

// import "os/exec"
// import "os"
// import "path/filepath"
// import "time"
// import "regexp"
// import "strings"

// var watched map[string]int64 = make(map[string]int64)
// var cmd *exec.Cmd

// /*
// todo: your app's deploy config.json should mention all the commands to run and watch folders for dev.
// guppy is an open source tool for developing and deploying your code to highf.in
// */

// func Watch() {
// 	c := time.Tick(1 * time.Second)
// 	for _ = range c {
// 		if IsChanged() == true {
// 			go RunSalmon()
// 		}
// 	}
// }

// func RunSalmon() {
// 	if cmd != nil {
// 		Log("Change detected...")
// 		cmd.Process.Kill()
// 		cmd = nil
// 	}
// 	// todo: conditional
// 	//_ = exec.Command("go", "get", "code.highf.in/chalkhq/salmon").Run()

// 	// cmd = exec.Command("go", "install", "code.highf.in/chalkhq/tail")
// 	// cmd.Stdout = os.Stdout
// 	// cmd.Stderr = os.Stderr
// 	// err := cmd.Run()
// 	err := E("go install code.highf.in/chalkhq/tail").Run()
// 	if err != nil {
// 		Log("Attempting to run last known working go binary for Tail")
// 	} else {
// 		Log("Running tail..")
// 	}

// 	// todo: cmd should be a channel
// 	_ = E("/vagrant/salmon/go/bin/tail").Run()
// 	// cmd.Stdout = os.Stdout
// 	// cmd.Stderr = os.Stderr

// 	// _ = cmd.Run()
// }
