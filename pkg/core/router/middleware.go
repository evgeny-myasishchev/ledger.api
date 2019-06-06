package router

import "net/http"

// CreateCorsMiddlewareFunc - creates a middleware to handle CORS preflights
// TODO: Unit test
func CreateCorsMiddlewareFunc() func(next http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, req *http.Request) {
			// TODO: Make origins configurable
			// and restrict them for prod
			w.Header().Add("Access-Control-Allow-Origin", "*")
			w.Header().Add("Access-Control-Allow-Headers", "X-Request-ID,Authorization")
			if req.Method == "OPTIONS" {
				w.WriteHeader(200)
			} else {
				next(w, req)
			}
		}
	}
}
