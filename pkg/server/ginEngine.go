package server

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"ledger.api/pkg/logging"
)

type ginEngine struct {
	ginRouter *gin.Engine
}

func (engine *ginEngine) Handle(method string, path string, handler http.HandlerFunc) {
	engine.ginRouter.Handle(method, path, func(context *gin.Context) {
		handler.ServeHTTP(context.Writer, context.Request)
	})
}

func (engine *ginEngine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	engine.ginRouter.ServeHTTP(w, req)
}

func (engine *ginEngine) Run(port string) error {
	return engine.ginRouter.Run(port)
}

const noRouteErrFmt = `{ "errors": [ { "status": "%v", "title": "%v" } ] }`

var noRouteErrorBody = []byte(fmt.Sprintf(noRouteErrFmt, http.StatusNotFound, http.StatusText(http.StatusNotFound)))

func createGinEngine(logger logging.Logger) HTTPEngine {

	gin.DisableConsoleColor()
	gin.SetMode(gin.TestMode)
	ginRouter := gin.New()

	ginRouter.NoRoute(func(c *gin.Context) {
		c.Status(http.StatusNotFound)
		c.Writer.Write(noRouteErrorBody)
	})
	return &ginEngine{ginRouter: ginRouter}
}
