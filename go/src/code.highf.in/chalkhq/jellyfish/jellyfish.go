package main

import (
	"code.highf.in/chalkhq/shared/config"

	//"code.highf.in/chalkhq/shared/types"
	//"code.highf.in/chalkhq/shared/nodejs"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"os/exec"
	//"strconv"
	//"strings"
	//"strings"
)

const squid_port = ":8081"

var jellyports map[string]string

// type JellyProxy struct{}

// func (jellyProxy *JellyProxy) ServeHTTP(w http.ResponseWriter, req *http.Request) {
// 	// reverse proxy based on req.HOST to sharkport
// 	jellyport := strings.SplitAfter(req.Host, ":")[1]
// 	Log("proxy-jellyport===" + jellyport)
// 	sharkport := jellyports[jellyport]

// 	Log("proxy-sharkport===" + sharkport)
// 	Log("things: " + req.Host + "/ /" + req.RequestURI)

// 	director := func(target *http.Request) {
// 		target.URL.Scheme = "http" // todo: change to https between containers
// 		target.URL.Host = sharkport
// 		target.URL.Path = req.URL.Path
// 		target.URL.RawQuery = req.URL.RawQuery
// 	}

// 	p := httputil.ReverseProxy{Director: director}

// 	p.ServeHTTP(w, req)

// }

func main() {
	fmt.Println("starting jellyfish...")
	//mesh := make(map[string]string)

	jellyports = make(map[string]string)

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

	//instance_name := string(os.Args[2]) // for announcing itself to other jellyfish

	http.HandleFunc("/___api", func(w http.ResponseWriter, r *http.Request) {
		// reverse proxy based on req.HOST to sharkport

		sharkports := r.FormValue("sharkports")

		fmt.Println(sharkports)

		err := json.Unmarshal([]byte(sharkports), &jellyports)

		if err != nil {
			fmt.Println("failed to unmarshal response")
			fmt.Println(err)
		}
		for jellyport := range jellyports {
			//sharkport := jellyports[jellyport]
			Log("jellyfish: listening on " + jellyport)
			//go http.ListenAndServe("127.0.0.1:"+jellyport, &JellyProxy{})

			//go http.ListenAndServe(":"+jellyport, &JellyProxy{})
			go proxy(":"+jellyport, jellyports[jellyport])

		}

		// for jellyport := range jellyports {
		// 	sharkport := jellyports[jellyport]
		// 	Log("proxyo-jelloport:" + sharkport + "//" + jellyport)
		// 	// todo: what is the significance of 10.10.10.99 range, get it programmatically
		// 	err = exec.Command(`iptables`, `-t`, `nat`, `-A`, `OUTPUT`, `-p`, `tcp`, `-d`, `10.10.10.99`, `--dport`, jellyport, `-j`, `DNAT`, `--to`, sharkport).Run()
		// 	//err = exec.Command(`/sbin/iptables`, `-t`, `nat`, `-A`, `OUTPUT`, `-p`, `tcp`, `-d`, `10.10.10.99`, `--dport`, `6000`, `-j`, `DNAT`, `--to`, `10.10.10.11:49225`).Run()

		// 	// iptables -t nat -A OUTPUT -p tcp -d 10.10.10.99 --dport 6100 -j DNAT --to 10.10.10.11:49234
		// 	// iptables -t nat -A OUTPUT -p tcp -d 10.10.10.99 --dport 6000 -j DNAT --to 10.10.10.11:49236
		// 	if err != nil {
		// 		Log("error: " + err.Error())
		// 	}

		// }

	})

	// execute all the parts of the app
	for k := 0; k < len(app.Execs); k++ {
		appPart := app.Execs[k]

		if appPart.Lang == "mongodb" {
			fmt.Println("is mongo")
			defer func() {
				proxy("0.0.0.0"+squid_port, "127.0.0.1:27017")

			}()
		} else {

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
			defer func() {
				http.ListenAndServe(squid_port, nil)
				fmt.Println("Listening on " + squid_port)
			}()
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

		case "golang":
			fmt.Println("running app..")
			// todo should support command line params
			cmd := exec.Command(appPart.Main) //+app.Main)
			//go cmd.Run()
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			go cmd.Run()
		case "mongodb":
			cmd := exec.Command("/usr/bin/mongod", "--smallfiles") //+app.Main)
			//go cmd.Run()
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			go cmd.Run()
		}
	}

	// static file handlers
	for i := 0; i < len(app.Statics); i++ {
		static := app.Statics[i]
		http.Handle(static.Path, http.FileServer(http.Dir("/code/"+static.Dir)))

		//mesh[static.Path] = static.Dir
	}

}

func Log(msg string) {
	// very temporary logging utility
	_, _ = http.Get("http://10.10.10.5/" + msg)

}

func forward(local net.Conn, remoteAddr string) {
	remote, err := net.Dial("tcp", remoteAddr)
	if remote == nil {
		fmt.Fprintf(os.Stderr, "remote dial failed: %v\n", err)
		return
	}
	go io.Copy(local, remote)
	go io.Copy(remote, local)
}

func proxy(localAddr string, remoteAddr string) {

	local, _ := net.Listen("tcp", localAddr)
	if local == nil {
		//fatal("cannot listen: %v", err)
	}
	for {
		conn, err := local.Accept()
		if err != nil {
			continue
		}
		go forward(conn, remoteAddr)
	}
}
