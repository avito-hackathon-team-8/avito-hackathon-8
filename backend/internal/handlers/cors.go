package handlers

import "net/http"

var localOrigins = map[string]struct{}{
	"http://localhost:3000": {},
	"http://localhost:5173": {},
	"http://127.0.0.1:3000": {},
	"http://127.0.0.1:5173": {},
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		origin := request.Header.Get("Origin")
		if _, allowed := localOrigins[origin]; allowed {
			response.Header().Set("Access-Control-Allow-Origin", origin)
			response.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
			response.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, OPTIONS")
			response.Header().Add("Vary", "Origin")

			if request.Method == http.MethodOptions {
				response.WriteHeader(http.StatusNoContent)

				return
			}
		}

		next.ServeHTTP(response, request)
	})
}
