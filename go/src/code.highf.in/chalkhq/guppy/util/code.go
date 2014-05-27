package util

import (
	"code.highf.in/chalkhq/highfin/config"
	"code.highf.in/chalkhq/highfin/nodejs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var dashConfig config.DashConfig
var watched map[string]int64 = make(map[string]int64)

//var cmd *exec.Cmd
var cmds []*exec.Cmd

func Run() {
	dashConfig = config.GetDashConfig("./")

	c := time.Tick(1 * time.Second)
	for _ = range c {
		// watch/run all apps unless specific app is provided in commandline
		if len(os.Args) >= 3 {
			// todo: this check should be an external function, maybe a method on dashConfig struct
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

func isChanged(app config.App) bool {
	//Log("Hello Guppy Watch util\n")
	// make it walk the folder tree.. if it's fast enough we can just hash it every few seconds and compare, or count the number of files.
	// or even better start using https://github.com/howeyc/fsnotify which will become an official api in go1.4
	new_watched := make(map[string]int64)
	changed := false
	// walk the src folder hashing every file and appending the hash to a string
	// hash the string
	// store in memory, repeat and compare
	for k := 0; k < len(app.Execs); k++ {
		appPart := app.Execs[k]

		for i := 0; i < len(appPart.Watch); i++ {

			_ = filepath.Walk(appPart.Watch[i], func(path string, info os.FileInfo, err error) error {

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
					//Log(path)
					changed = true
				}

				// move it from old path to new path
				delete(watched, path)
				new_watched[path] = time

				return nil

			})
		}
	}

	// if we didn't encounter paths we previously watched
	if len(watched) > 0 {
		changed = true
		Log("Change detected...")
	}

	watched = new_watched
	return changed
}

// todo: should run all the commands in exec. refactor E() function anyway to support passing in the command object
func runApp(app config.App) {
	for i := 0; i < len(cmds); i++ {
		cmds[i].Process.Kill()
	}

	cmds = make([]*exec.Cmd, len(app.Execs))

	for i := 0; i < len(app.Execs); i++ {
		appPart := app.Execs[i]
		switch appPart.Lang {
		case "golang":
			Log("Installing")

			err := E("go install " + appPart.Main).Run()
			if err != nil {
				Log("failed to install app")
				Log("Re-running last successful build..")
			} else {
				Log("Running..")
			}
			mainSplit := strings.Split(appPart.Main, "/")
			cmds[i] = E(mainSplit[len(mainSplit)-1])
			cmds[i].Run()

		case "nodejs":
			Log("Running..")
			nodejs.InstallNode(appPart.Version)
			cmds[i] = E(nodejs.BinPath(appPart.Version) + ` ` + appPart.Main)
			cmds[i].Run()
		}
	}
}

func NpmInstall() {
	// godep go install code.highf.in/chalkhq/highfin
	dashConfig = config.GetDashConfig("./")
	if app, ok := dashConfig.Apps[os.Args[2]]; ok == true {
		for i := 0; i < len(app.Execs); i++ {
			appPart := app.Execs[i]
			switch appPart.Lang {
			case "nodejs":
				for i := range appPart.Npm {
					path := dashConfig.BasePath + `/` + appPart.Npm[i]
					err := E(nodejs.NpmPath(appPart.Version) + ` --prefix ` + path + ` install ` + path).Run()
					if err != nil {
						Log(`failed to npm install ` + path)
					}

				}
			}
		}
	} else {
		Log(`App "` + os.Args[2] + `" is not declared in -.json`)
		return
	}
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

	_cmd := exec.Command(command_string)
	_cmd.Args = args
	_cmd.Stdout = os.Stdout
	_cmd.Stderr = os.Stderr

	return _cmd

}
