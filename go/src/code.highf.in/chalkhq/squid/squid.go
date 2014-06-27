package main

/*
Squid is a reverse proxy and load balancer
*/

import (
	"code.highf.in/chalkhq/shared/persistence"
	"code.highf.in/chalkhq/shared/types"
	//"code.highf.in/chalkhq/squid/api"
	"encoding/json"
	"fmt"
	//"math"

	"net/http"
	"net/http/httputil"

	"strings"
	//"net/url"
	//"net/http/httputil"
)

type Domain struct {
	Request_count int
	Bindex        int // round-robbin / load balance index
	Sharkports    []string
	Env           string // account_project_env
}

// map deploy id to domains, for easy pulldown
var idMap map[string][]string

var domainMap map[string]Domain

type Api struct{}

var changed bool

func (api *Api) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	fmt.Println("/")
	r := types.Response{}

	r.Response.Meta.Status = 200

	r.Response.Meta.Errors = make([]string, 0)

	r.Response.Data = make(map[string]interface{})

	r.W = w

	r.Req = *req
	r.Req.ParseForm()
	//r.Req.ParseMultipartForm(64)

	r.Fields = r.Req.Form
	if len(r.Hashbang) > 0 {
		r.Segments = strings.Split(r.Hashbang, "/")
	}

	router(r)
}

func main() {
	fmt.Println("starting...")

	changed = false

	domainMap = make(map[string]Domain)
	go persistence.GetData(&domainMap, "/squid/domainMap.json")
	go persistence.PersistData(&domainMap, "/squid/domainMap.json", &changed)
	//domainMap := make(map[string]string)

	//domainMap["app1.test"] = "10.10.10.11:5000"

	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		//fmt.Println(string(req.Host + req.URL.Path))

		//fmt.Println(".....", len(domainMap[req.Host].Sharkports))

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
			bindex := domainMap[req.Host].Bindex
			//domainMap[req.Host].Bindex = int(math.Mod(float64(domainMap[req.Host].Bindex+1), float64(len(domainMap[req.Host].Sharkports))))
			//domainMap[req.Host].Request_count++
			//fmt.Println(domainMap[req.Host].Sharkports[bindex])
			target.URL.Scheme = "http"
			target.URL.Host = domainMap[req.Host].Sharkports[bindex]
			target.URL.Path = req.URL.Path
			target.URL.RawQuery = req.URL.RawQuery
		}

		p := httputil.ReverseProxy{Director: director}

		p.ServeHTTP(w, req)

	})

	go http.ListenAndServe(":80", nil)

	fmt.Println("Listening on 80")

	http.ListenAndServe(":8282", &Api{})

	fmt.Println("Listening on :8282 for internal api")

}

func router(r types.Response) {
	fmt.Println("router")
	switch r.Req.URL.Path {
	case "/route/update":
		fmt.Println("update called")
		//api.Route_Create(r, &domainMap)
		updateRoute(r)
		defer func() { changed = true }()

	default:
		r.Response.Meta.Status = 404
		r.AddError("Path: " + r.Req.URL.Path + " is not valid")
		r.Kill(200)
	}
}

func updateRoute(r types.Response) {
	account := r.Req.FormValue("account")
	project := r.Req.FormValue("project")
	branch := r.Req.FormValue("branch")
	domains := r.Req.FormValue("domains")

	fmt.Println(domains)

	env := account + "_" + project + "_" + branch

	var domainMappings map[string][]string

	err := json.Unmarshal([]byte(domains), &domainMappings)

	if err != nil {
		fmt.Println("failed to unmarshal response")
		fmt.Println(err)
	}

	for domain := range domainMappings {
		if _, ok := domainMap[domain]; ok == true && domainMap[domain].Env != env {
			// abort, this domain is in use by another user
			// todo: cascade error down to user
			r.AddError("domain already in use by another user " + domain)
			r.Kill(500)
			return
		}

		fmt.Println(domain)
		fmt.Println(domainMappings[domain][0])
		domainMap[domain] = Domain{
			0, 0,
			domainMappings[domain],
			env}

	}

	//todo: for now just simply replace it
	// todo: should have logic to check if domain is being used by another user

}
