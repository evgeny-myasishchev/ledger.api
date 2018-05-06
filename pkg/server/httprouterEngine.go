package server

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"ledger.api/pkg/logging"
)

type httpRouterEngine struct {
	router *httprouter.Router
}

func (engine *httpRouterEngine) Handle(method string, path string, handler http.HandlerFunc) {
	engine.router.Handle(method, path, func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		handler.ServeHTTP(w, r)
	})
}

func (engine *httpRouterEngine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	engine.router.ServeHTTP(w, req)
}

func (engine *httpRouterEngine) Run(port string) error {
	return http.ListenAndServe(port, engine.router)
}

func createHTTPRouterEngine(logger logging.Logger) HTTPEngine {
	router := httprouter.New()
	router.NotFound = func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write(noRouteErrorBody)
	}
	return &httpRouterEngine{router: router}
}
