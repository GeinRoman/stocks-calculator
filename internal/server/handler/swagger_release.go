//go:build !debug

package handler

import "net/http"

func registerSwagger(mux *http.ServeMux) {}
