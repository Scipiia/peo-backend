package auth

import (
	"encoding/base64"
	"net/http"
	"strings"
)

func BasicAuth(username, password string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				requireAuth(w)
				return
			}

			if !strings.HasPrefix(authHeader, "Basic ") {
				requireAuth(w)
				return
			}

			creds, err := base64.StdEncoding.DecodeString(authHeader[6:])
			if err != nil {
				requireAuth(w)
				return
			}

			credPair := strings.SplitN(string(creds), ":", 2)
			if len(credPair) != 2 {
				requireAuth(w)
				return
			}

			if credPair[0] != username || credPair[1] != password {
				requireAuth(w)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func requireAuth(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", `Basic realm="Admin Area"`)
	http.Error(w, "Unauthorized", http.StatusUnauthorized)
}
