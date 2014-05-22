package util

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var dashConfig DashConfig
var watched map[string]int64 = make(map[string]int64)
var cmd *exec.Cmd

func Run() {
	dashConfig = GetDashConfig()

	c := time.Tick(1 * time.Second)
	for _ = range c {
		// watch/run all apps unless specific app is provided in commandline
		if len(os.Args) >= 3 {
			if app, ok := dashConfig.Apps[os.Args[2]]; ok == true {
				if isChanged(app) == true {
					go runApp(app)
				}
			} else {
				Log(`App "` + os.Args[2] + `" is not declared in -.json`)
				return
			}
		} else {
			for j := range dashConfig.Apps {
				app := dashConfig.Apps[j]
				if isChanged(app) == true {
					go runApp(app)
				}
			}
		}
	}
}

func isChanged(app App) bool {
	//Log("Hello Guppy Watch util\n")
	// make it walk the folder tree.. if it's fast enough we can just hash it every few seconds and compare, or count the number of files.
	// or even better start using https://github.com/howeyc/fsnotify which will become an official api in go1.4
	new_watched := make(map[string]int64)
	changed := false
	// walk the src folder hashing every file and appending the hash to a string
	// hash the string
	// store in memory, repeat and compare

	for i := 0; i < len(app.Watch); i++ {

		_ = filepath.Walk(app.Watch[i], func(path string, info os.FileInfo, err error) error {

			if err != nil {
				Log("Can't watch " + path + ", path does not exist")
				return nil
			}

			// todo(yoav): import path should be: "/vagrant/go/src/code.highf.in/chalkhq/salmon", then salmon can reuse the same functions without special treatment
			// however, then github will only be a mirror, and you'll be pushing/pulling from shark.
			// given that shark is not stable yet, this will have to suffice.
			ignored_paths := regexp.MustCompile(`/.git/`)

			if ignored_paths.MatchString(path) {
				return nil
			}

			time := info.ModTime().UnixNano()
			// if it's a new or modified path

			if _, i := watched[path]; i == false || watched[path] != time {
				// new file to watch
				changed = true
			}

			// move it from old path to new path
			delete(watched, path)
			new_watched[path] = time

			return nil

		})
	}

	// if we didn't encounter paths we previously watched
	if len(watched) > 0 {
		changed = true
	}

	watched = new_watched
	return changed
}

func runApp(app App) {
	if cmd != nil {
		Log("Change detected...")
		cmd.Process.Kill()
		cmd = nil
	}
	switch app.Lang {
	case "golang":
		err := E("go install " + app.Main).Run()
		if err != nil {
			Log("failed to install app")
		}
		mainSplit := strings.Split(app.Main, "/")
		err = E(mainSplit[len(mainSplit)-1]).Run()
		if err != nil {
			//Log("failed to run app", err.Error())
		}
	case "nodejs":
		// this should just use the function directly instead of depending on another guppy process
		err := E("guppy node " + app.Version + " " + app.Main).Run()
		if err != nil {
			Log("failed to install app")
		}
	}
}

func Build() {
	// godep go install code.highf.in/chalkhq/highfin
}

func CompileLESS() {

}

// wraps the exec.Command() function automatically attaching stdout/err and splitting a single string command/args into their appropriate pieces
func E(command string) *exec.Cmd {
	args := strings.Split(command, ` `)
	command_string := strings.Trim(args[0], ` `)
	for i := 0; i < len(args); i++ {
		args[i] = strings.Trim(args[i], ` `)
	}

	cmd = exec.Command(command_string)
	cmd.Args = args
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd

}
