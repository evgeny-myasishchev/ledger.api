package app

import (
	"encoding/json"
	"net/http"

	"ledger.api/pkg/core/router"
)

func createHealthcheckPingHandler() router.ToolkitHandlerFunc {
	pingResponse, err := json.Marshal(map[string]interface{}{
		"ping": "PONG",
	})
	if err != nil {
		panic(err)
	}

	return func(w http.ResponseWriter, r *http.Request, h router.HandlerToolkit) error {
		w.Header().Add("content-type", "application/json")
		_, err := w.Write(pingResponse)
		return err
	}
}

// func createHealthcheckInfoHandler() router.ToolkitHandlerFunc {
// 	clusterName := os.Getenv("CLUSTER_NAME")
// 	infoResponse, err := json.Marshal(map[string]interface{}{
// 		"appName":     version.AppName,
// 		"appVersion":  version.VERSION,
// 		"clusterName": clusterName,
// 		"git": map[string]interface{}{
// 			"hash": version.GitHash,
// 			"ref":  version.GitRef,
// 			"url":  version.GitURL,
// 		},
// 	})
// 	if err != nil {
// 		panic(err)
// 	}

// 	return func(w http.ResponseWriter, r *http.Request, h router.HandlerToolkit) error {
// 		w.Header().Add("content-type", "application/json")
// 		_, err := w.Write(infoResponse)
// 		return err
// 	}
// }

// SetupRoutes will register HC related routes
func SetupRoutes(appRouter router.Router) {
	appRouter.Handle("GET", "/v2/healthcheck/ping", createHealthcheckPingHandler())

	// TODO: Restore info route
	// appRouter.Handle("GET", "/v2/healthcheck/info", createHealthcheckInfoHandler())
}
