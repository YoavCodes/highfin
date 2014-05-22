package main

/*
Octopus is a Shark & Sqid co-ordinator.
- spins up new instances
- aggregates and respondds to health checks
- maintains open connection to Squids and Sharks at all times
- allocates Sharks to deploy containers
- manages Coral and builds containers before sending tar to a given Shark to deploy
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

	domainMap["app1.test"] = "10.10.10.10:50000"

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
		fmt.Println("test2")
		director := func(target *http.Request) {
			target.URL.Scheme = "http"
			target.URL.Host = domainMap[req.Host]
			fmt.Println(".....", domainMap[req.Host])
			target.URL.Path = req.URL.Path
			target.URL.RawQuery = req.URL.RawQuery
		}

		p := httputil.ReverseProxy{Director: director}

		p.ServeHTTP(w, req)

	})
	http.ListenAndServe(":8080", nil)
	fmt.Println("Listening on 8080")
}
