package types

import (
	"encoding/json"
	"net/http"
)

type Response struct {
	Fields   map[string][]string
	Files    map[string]string
	Req      http.Request
	Headers  map[string]string
	Pathname string
	Hashbang string
	Command  string
	Segments []string
	W        http.ResponseWriter
	Response struct {
		Meta struct {
			Status int      `json:"status"`
			Errors []string `json:"errors"`
		} `json:"meta"`
		Data map[string]interface{} `json:"data"`
	}
}

func (r *Response) Kill(status int) {

	// return response
	r.W.Header().Set("Content-Type", "application/json")

	r.Response.Meta.Status = status

	res, _ := json.Marshal(r.Response)

	r.W.WriteHeader(200)

	r.W.Write(res)

}

func (r *Response) AddError(error_string string) {
	r.Response.Meta.Errors = append(r.Response.Meta.Errors, error_string)
}
