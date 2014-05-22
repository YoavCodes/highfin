package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"os/exec"
)

func main() {
	fmt.Println("starting jellyfish...")
	domainMap := make(map[string]string)

	domainMap["app1.test"] = "127.0.0.1:8080"
	// start app
	cmd := exec.Command("/code/test")
	go cmd.Run()

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
	fmt.Println("Listening on 8080")
}
