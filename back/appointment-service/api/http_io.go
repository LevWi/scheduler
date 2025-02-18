package server

import "net/http"

type HttpIO struct {
	Wrt http.ResponseWriter
	Req *http.Request
}
