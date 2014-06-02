package main

import (
	"code.highf.in/chalkhq/highfin/config"
	//"code.highf.in/chalkhq/highfin/nodejs"
	"fmt"
	"net/http"
	"net/http/httputil"
	"os"
	"os/exec"
	//"strings"
)

const squid_port = ":8081"

func main() {
	fmt.Println("starting jellyfish...")
	//mesh := make(map[string]string)

	// get config
	dashConfig := config.GetDashConfig("/code/")

	//mesh["app1.test"] = "127.0.0.1:8080"

	if len(os.Args) < 2 {
		return
	}

	/*c := exec.Command(`ls`, `/code/`)
	c.Stderr = os.Stderr
	c.Stdout = os.Stdout
	_ = c.Run()*/

	app_name := string(os.Args[1])
	app := dashConfig.Apps[app_name] // specified in the Dockerfile ENTRYPOINT line by octopus

	// execute all the parts of the app
	for k := 0; k < len(app.Execs); k++ {
		appPart := app.Execs[k]
		//fmt.Println("appPart " + string(k) + " len: " + string(len(appPart.Endpoints)) + " / " + appPart.Endpoints[0].Path)
		for i := 0; i < len(appPart.Endpoints); i++ {
			mapping := appPart.Endpoints[i]
			// mesh[app1.test] = 127.0.0.1:8080
			//mesh[mapping.Path] = "127.0.0.1:" + mapping.Port
			fmt.Println("adding path: " + mapping.Path + " - " + mapping.Port)
			http.HandleFunc(mapping.Path, func(w http.ResponseWriter, req *http.Request) {

				fmt.Println("jellyfish")
				fmt.Println(string(req.Host + req.URL.Path))
				director := func(target *http.Request) {
					target.URL.Scheme = "http"
					target.URL.Host = "127.0.0.1:" + mapping.Port
					target.URL.Path = req.URL.Path
					target.URL.RawQuery = req.URL.RawQuery
				}

				p := httputil.ReverseProxy{Director: director}

				p.ServeHTTP(w, req)

			})
		}

		// install app, should happen once
		switch appPart.Lang {
		case "nodejs":
			// note: I am strongly against running npm install on the server. npm install should be run locally in fry-box.
			// a network error or dependency on github being unaccessable at some arbitrary time during a redeploy or rebuild on the server
			// should never ever ever be responsible for deploy issues. commit ALL dependecies to your git repo so your app just works.

			// for i := range app.Npm {
			// 	path := `/code/` + app.Npm[i]

			// 	cmd := exec.Command("code/__dep/n/"+app.Version+"/bin/npm", "--prefix", path, "install", path)

			// 	cmd.Stderr = os.Stderr
			// 	cmd.Stdout = os.Stdout
			// 	err := cmd.Run()
			// 	if err != nil {
			// 		fmt.Println(`failed to npm install ` + err.Error() + path + "::: ")
			// 	}

			// }
		}

		// todo: for every folder in app.Public, look for the pathname and try serve. if none found then serve 404.
		// should load all public folder files into memory on start, a'la memcachy.
		// hmmm, in salmon we do this with routing, we could still. -.json should map pathname to folder. by default / maps to /code/public/
		// we also have to map /0 to the node.js app then. we want to check for /0, then look for static file, then serve 404.
		// -.json should contain url mappings: salmon: {endpoint: http://domain.com/0, public: {http://domain.com/: /public})
		// this way user could map their root domain to node.js endpoint and have their app deal with static/public content.

		// run app, should happen initially and whenever the app exits, with time delays if it exits more than once per second or whatever
		switch appPart.Lang {
		case "nodejs":
			fmt.Println("running app..")
			cmd := exec.Command("/code/__dep/n/"+appPart.Version+"/bin/node", "/code/"+appPart.Main) //+app.Main)
			//go cmd.Run()
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			go cmd.Run()
			// if err != nil {
			// 	fmt.Println(err.Error())
			// }
		}
	}

	// proxy

	// http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
	// 	fmt.Println("jellyfish")
	// 	fmt.Println(string(req.Host + req.URL.Path))
	// 	if _, ok := domainMap[req.Host]; ok == false {
	// 		// domain mapping doesn't exist, return 404

	// 		scheme := "http"
	// 		if req.TLS != nil {
	// 			scheme += "s"
	// 		}

	// 		w.Write([]byte(`<a href="http://highf.in">HighF.in</a> 404: Sorry ` + scheme + "://" + req.Host + ` doesn't exist. bloop bloop`))
	// 		return
	// 	}
	// 	fmt.Println("test2")
	// 	director := func(target *http.Request) {
	// 		target.URL.Scheme = "http"
	// 		target.URL.Host = domainMap[req.Host]
	// 		target.URL.Path = req.URL.Path
	// 		target.URL.RawQuery = req.URL.RawQuery
	// 	}

	// 	p := httputil.ReverseProxy{Director: director}

	// 	p.ServeHTTP(w, req)

	// })

	// static file handlers
	for i := 0; i < len(app.Statics); i++ {
		static := app.Statics[i]
		http.Handle(static.Path, http.FileServer(http.Dir("/code/"+static.Dir)))

		//mesh[static.Path] = static.Dir
	}

	http.ListenAndServe(squid_port, nil)
	fmt.Println("Listening on " + squid_port)
}
