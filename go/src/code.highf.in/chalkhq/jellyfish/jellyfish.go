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

func main() {
	fmt.Println("starting jellyfish...")
	domainMap := make(map[string]string)

	domainMap["app1.test"] = "127.0.0.1:8080"
	// get config
	dashConfig := config.GetDashConfig("/code/")

	if len(os.Args) < 2 {
		return
	}

	/*c := exec.Command(`ls`, `/code/`)
	c.Stderr = os.Stderr
	c.Stdout = os.Stdout
	_ = c.Run()*/

	app_name := string(os.Args[1])
	app := dashConfig.Apps[app_name] // specified in the Dockerfile ENTRYPOINT line by octopus

	// install app, should happen once
	switch app.Lang {
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

	// run app, should happen initially and whenever the app exits, with time delays if it exits more than once per second or whatever
	switch app.Lang {
	case "nodejs":
		fmt.Println("running app..")
		cmd := exec.Command("/code/__dep/n/"+app.Version+"/bin/node", "/code/salmon.js") //+app.Main)
		//go cmd.Run()
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		go cmd.Run()
		// if err != nil {
		// 	fmt.Println(err.Error())
		// }
	}

	// proxy
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		fmt.Println("jellyfish")
		fmt.Println(string(req.Host + req.URL.Path))
		if _, ok := domainMap[req.Host]; ok == false {
			// domain mapping doesn't exist, return 404

			scheme := "http"
			if req.TLS != nil {
				scheme += "s"
			}

			w.Write([]byte(`<a href="http://highf.in">HighF.in</a> 404: Sorry ` + scheme + "://" + req.Host + ` doesn't exist. bloop bloop`))
			return
		}
		fmt.Println("test2")
		director := func(target *http.Request) {
			target.URL.Scheme = "http"
			target.URL.Host = domainMap[req.Host]
			target.URL.Path = req.URL.Path
			target.URL.RawQuery = req.URL.RawQuery
		}

		p := httputil.ReverseProxy{Director: director}

		p.ServeHTTP(w, req)

	})
	http.ListenAndServe(":8081", nil)
	fmt.Println("Listening on 8081")
}
