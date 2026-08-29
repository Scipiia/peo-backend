package auth

import (
	"context"
	"encoding/base64"
	"net/http"
	"strings"
	"vue-golang/internal/auth-ldap"
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

func JWTAuth(jwtService *auth_ldap.JWTService) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, `{"error": "Authorization header required"}`, http.StatusUnauthorized)
				return
			}

			// Ожидаем формат "Bearer <token>"
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == authHeader {
				http.Error(w, `{"error": "Invalid authorization format. Use 'Bearer <token>'"}`, http.StatusUnauthorized)
				return
			}

			claims, err := jwtService.ValidateToken(tokenString)
			if err != nil {
				http.Error(w, `{"error": "Invalid or expired token"}`, http.StatusUnauthorized)
				return
			}

			// Кладем claims в контекст запроса, чтобы хэндлеры могли их читать
			ctx := context.WithValue(r.Context(), "user_claims", claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequirePermission(permission string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := r.Context().Value("user_claims").(*auth_ldap.CustomClaims)
			if !ok {
				http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
				return
			}

			// Ищем нужное право в списке прав пользователя
			for _, p := range claims.Permissions {
				if p == permission {
					next.ServeHTTP(w, r) // Право есть, пропускаем дальше
					return
				}
			}

			http.Error(w, `{"error": "Forbidden: missing permission `+permission+`"}`, http.StatusForbidden)
		})
	}
}
