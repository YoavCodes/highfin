package util

import (
	"code.highf.in/chalkhq/shared/command"
	"code.highf.in/chalkhq/shared/config"
	"code.highf.in/chalkhq/shared/log"
	"code.highf.in/chalkhq/shared/nodejs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
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
				log.Log(`App "` + os.Args[2] + `" is not declared in -.json`)
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
	//log.Log("Hello Guppy Watch util\n")
	// make it walk the folder tree.. if it's fast enough we can just hash it every few seconds and compare, or count the number of files.
	// or even better start using https://github.com/howeyc/fsnotify which will become an official api in go1.4
	new_watched := make(map[string]int64)
	changed := false
	less_changed := false
	// walk the src folder hashing every file and appending the hash to a string
	// hash the string
	// store in memory, repeat and compare

	for k := 0; k < len(app.Execs); k++ {
		appPart := app.Execs[k]

		var ignored_path_regex *regexp.Regexp
		var less_path_regex *regexp.Regexp
		// todo(yoav): import path should be: "/vagrant/go/src/code.highf.in/chalkhq/salmon", then salmon can reuse the same functions without special treatment
		// however, then github will only be a mirror, and you'll be pushing/pulling from shark.
		// given that shark is not stable yet, this will have to suffice.
		if len(appPart.Exclude) > 0 {
			ignored_paths := strings.Join(appPart.Exclude, "*|") + "*"
			ignored_path_regex = regexp.MustCompile(ignored_paths)
		}
		less_paths := ""
		for i := 0; i < len(appPart.Less); i++ {
			less_paths += appPart.Less[i].From + "*|"
		}

		if len(less_paths) > 0 {
			less_paths = less_paths[:len(less_paths)-2] // get rid of final |
			less_path_regex = regexp.MustCompile(less_paths)
		}
		for i := 0; i < len(appPart.Watch); i++ {

			_ = filepath.Walk(appPart.Watch[i], func(path string, info os.FileInfo, err error) error {

				if err != nil {
					log.Log("Can't watch " + path + ", path does not exist")
					log.LogE(err)
					return nil
				}

				if len(appPart.Exclude) > 0 {
					if ignored_path_regex.MatchString(path) {
						if info.IsDir() {
							return filepath.SkipDir
						}
						return nil
					}
				}

				time := info.ModTime().UnixNano()
				// if it's a new or modified path
				if changed == false {
					if _, i := watched[path]; i == false || watched[path] != time {
						// new file to watch
						if info.IsDir() == false {
							changed = true
							//log.Log(path)
						}
					}
				}

				if _, i := watched[path]; i == false || watched[path] != time {
					if len(less_paths) > 0 {
						if less_path_regex.MatchString(path) {
							less_changed = true
						}
					}
				}

				// move it from old path to new path
				delete(watched, path)
				new_watched[path] = time

				return nil

			})
		}

		if less_changed == true {
			changed = compileLESS(appPart)
			// we want execution flow to wait for less to compile
		}
	}

	// if we didn't encounter paths we previously watched
	if len(watched) > 0 {
		changed = true
		log.Log("Change detected...")
	}

	watched = new_watched
	return changed
}

// todo: should run all the commands in exec. refactor command.E() function anyway to support passing in the command object
func runApp(app config.App) {
	for i := 0; i < len(cmds); i++ {
		// todo: should check to see if process is still running first
		if cmds[i] != nil {
			// give process a chance to gracefully shutdown
			// we want to send all the sigterms asap and then give it half a second to shutdown
			cmds[i].Process.Signal(syscall.SIGTERM)
			cmd := cmds[i]
			go func(cmd *exec.Cmd) {
				// note: i will exist in scope, but it's value would be changed by the for loop by the time this executes
				// note: fishtank only needs around <40 ms to terminate gracefully. 200 should give more than enough headroom
				// and still be snappy for the user.
				time.Sleep(400 * time.Millisecond)
				cmd.Process.Kill()
			}(cmd)
		}
	}
	// let all the sigterm delays expire and proccesses Kill()ed before continuing
	time.Sleep(400 * time.Millisecond)
	// find a better way to get the real number of execs including grunt processes
	cmds = make([]*exec.Cmd, len(app.Execs))

	for i := 0; i < len(app.Execs); i++ {
		appPart := app.Execs[i]

		switch appPart.Lang {
		case "golang":
			log.Log("Installing")

			err := command.E("go install " + appPart.Main).Run()
			if err != nil {
				log.Log("failed to install app")
				log.Log("Re-running last successful build..")
			} else {
				log.Log("Running..")
			}
			mainSplit := strings.Split(appPart.Main, "/")
			golang := command.E(mainSplit[len(mainSplit)-1])
			cmds[i] = golang
			golang.Run()

		case "nodejs":
			/*
				// disable executing grunt for now.
				// note: required cmds = make() up there to make room for the grunt cmds
				if appPart.GruntDirectory != "" {
					log.Log("Running grunt..")
					grunt := command.E(nodejs.BinPath(appPart.Version) + ` ` + nodejs.GruntPath(appPart.Version) + ` dev`)
					gruntpath, _ := filepath.Abs(dashConfig.BasePath + `/` + appPart.GruntDirectory)
					grunt.Dir = gruntpath
					cmds[i] = grunt
					go grunt.Run()
					i++
				}*/
			runJasmine(appPart)
			log.Log("Running nodejs..")
			nodejs.InstallNode(appPart.Version)
			mainpath, _ := filepath.Abs(dashConfig.BasePath + `/` + appPart.Main)
			node := command.E(nodejs.BinPath(appPart.Version) + ` ` + mainpath)
			node.Dir, _ = filepath.Split(mainpath)
			cmds[i] = node
			node.Run()

		}
	}
}

func compileLESS(appPart config.Exec) bool {
	if len(appPart.Less) == 0 {
		return false
	}

	log.Log("Compiling less...")

	for i := 0; i < len(appPart.Less); i++ {
		less := appPart.Less[i]
		// loop over less folder and output into css folder
		_ = filepath.Walk(less.From, func(path string, info os.FileInfo, err error) error {

			if err != nil {
				log.Log("Can't find less, " + path + " does not exist")
				return nil
			}

			//source := path // asbolute path of less file
			destination_file := ""
			if path != less.From {
				destination_file = path[len(less.From):] //strings.Replace(path, less.From, ``, 1) // less file path minus the root less folder
			}
			//log.Log(path + " :: " + destination_file + " :: " + less.From)
			// note: less.To is expected to be an abs path, resolved in dashconfig.go GetDashConfig()
			//log.Log(destination_file)
			//log.Log(path)
			destination := less.To + "/"

			var cmd *exec.Cmd

			if info.IsDir() == false && path[len(path)-5:] != ".less" {
				// copy non-less files and folders
				destination += destination_file
				//log.Log("cp " + path + " " + destination)
				cmd = command.E("cp " + path + " " + destination)
			} else if info.IsDir() == true {
				destination += destination_file
				//log.Log("mkdir -p " + destination)
				cmd = command.E("mkdir -p " + destination)
			} else {
				// compile less
				destination = less.To + "/" + strings.Replace(destination_file, `.less`, `.css`, 1) // replace less extensions extension
				options := ""
				if less.Min == true {
					options = "-x"
				}
				cmd = command.E(nodejs.BinPath(appPart.Version) + ` ` + nodejs.LessPath(appPart.Version) + ` ` + options + ` ` + path + ` ` + destination)
			}
			//log.Log(nodejs.BinPath(appPart.Version) + ` ` + nodejs.LessPath(appPart.Version) + ` ` + path + ` ` + destination)
			cmd.Run()

			return nil
		})

	}
	return false
}

func runJasmine(appPart config.Exec) {
	if appPart.Jasmine.Backend != "" {
		log.Log("Running Jasmine back-end tests")
		command.E(nodejs.BinPath(appPart.Version) + ` ` + nodejs.JasmineNodePath(appPart.Version) + ` --forceexit ` + dashConfig.BasePath + `/` + appPart.Jasmine.Backend).Run()
	}

	if appPart.Jasmine.Frontend != "" {
		log.Log("Running Jasmine front-end tests")
		command.E(nodejs.BinPath(appPart.Version) + ` ` + nodejs.PhantomjsPath(appPart.Version) + ` --web-security=false ` + dashConfig.BasePath + `/` + appPart.Jasmine.Frontend + `/../.jasmine/run-jasmine.js` + ` ` + dashConfig.BasePath + `/` + appPart.Jasmine.Frontend + `/../.jasmine/SpecRunner.html`).Run()
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
					err := command.E(nodejs.NpmPath(appPart.Version) + ` --prefix ` + path + ` install ` + path).Run()
					if err != nil {
						log.Log(`failed to npm install ` + path)
					}

				}
			}
		}
	} else {
		log.Log(`App "` + os.Args[2] + `" is not declared in -.json`)
		return
	}
}

func updateCacheControl() {
	// todo later
	/*
			original js code from salmon
			var timestamp = new Date().getTime();
		    loopOverFolders([tail.config._public_path, tail.config._template_path], function(file_name, root_path) {
		        // modify .html, .js, and .css files.
		        if (file_name.match(/.html$|.js$|.css$|.less$/)) {
		            var file_contents = fs.readFileSync(root_path + file_name, {
		                'encoding': 'utf8'
		            });
		            file_contents = file_contents.replace(/_CACHE_CONTROL_=[0-9]+/g, "_CACHE_CONTROL_=" + timestamp);
		            fs.writeFileSync(root_path + file_name, file_contents, {
		                'encoding': 'utf8'
		            })
		        }
		    })
	*/
}
