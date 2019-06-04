package server

import (
	"context"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"ledger.api/pkg/logging"
)

type contextKey string

const requestParamsKey contextKey = "requestParams"

type httpRouterEngine struct {
	router *httprouter.Router
}

func (engine *httpRouterEngine) Handle(method string, path string, handler http.HandlerFunc) {
	engine.router.Handle(method, path, func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {

		contextWithParams := context.WithValue(r.Context(), requestParamsKey, params)
		reqWithParams := r.WithContext(contextWithParams)
		handler.ServeHTTP(w, reqWithParams)
	})
}

func (engine *httpRouterEngine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	engine.router.ServeHTTP(w, req)
}

func createHTTPRouterEngine(logger logging.Logger) HTTPEngine {
	router := httprouter.New()
	router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write(noRouteErrorBody)
	})
	return &httpRouterEngine{router: router}
}
