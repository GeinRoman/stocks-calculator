//go:build debug

package handler

import (
	"github.com/swaggo/http-swagger"
	"net/http"
	_ "stocks_calculator/docs"
)

func registerSwagger(mux *http.ServeMux) {
	mux.Handle("/swagger/", httpSwagger.WrapHandler)
}
