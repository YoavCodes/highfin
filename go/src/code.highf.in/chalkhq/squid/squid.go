package main

/*
Squid is a reverse proxy and load balancer
*/

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	//"net/url"
	//"net/http/httputil"
)

func main() {
	fmt.Println("starting...")
	domainMap := make(map[string]string)

	domainMap["app1.test"] = "10.10.10.11:5000"

	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		fmt.Println(string(req.Host + req.URL.Path))
		fmt.Println(".....", domainMap[req.Host])
		if _, ok := domainMap[req.Host]; ok == false {
			// domain mapping doesn't exist, return 404

			scheme := "http"
			if req.TLS != nil {
				scheme += "s"
			}

			w.Write([]byte(`<a href="http://highf.in">HighF.in</a> 404: Sorry ` + scheme + "://" + req.Host + ` doesn't exist. bloop bloop`))
			return
		}
		director := func(target *http.Request) {
			target.URL.Scheme = "http"
			target.URL.Host = domainMap[req.Host]
			target.URL.Path = req.URL.Path
			target.URL.RawQuery = req.URL.RawQuery
		}

		p := httputil.ReverseProxy{Director: director}

		p.ServeHTTP(w, req)

	})
	http.ListenAndServe(":80", nil)
	fmt.Println("Listening on 80")
}
