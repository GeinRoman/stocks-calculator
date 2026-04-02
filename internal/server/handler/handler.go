package handler

import (
	"net/http"
)

func New() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /login", login)
	registerSwagger(mux)
	return mux
}
