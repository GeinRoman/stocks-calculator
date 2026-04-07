//go:build debug

package handler

import (
	"github.com/swaggo/http-swagger"
	"net/http"
	_ "stocks_calculator/docs"
)


// @title           Stocks Calculator API
// @version         1.0
// @host            localhost:3333
// @BasePath        /
func registerSwagger(mux *http.ServeMux) {
	mux.Handle("/swagger/", httpSwagger.WrapHandler)
}
