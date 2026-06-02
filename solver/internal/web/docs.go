package web

import (
	_ "embed"
	"net/http"
)

//go:embed api/openapi.yaml
var openapiSpec []byte

func OpenAPIHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/yaml")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write(openapiSpec)
}

//go:embed api/scalar.html
var scalarHTML []byte

func DocsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(scalarHTML)
}
