package auth

import (
	"encoding/json"
	"log/slog"
	"net/http"
	auth_ldap "vue-golang/internal/auth-ldap"
)

type Authenticator interface {
	//AuthenticateUser(username, password string) (*auth_ldap.User, []string, error) // LDAP
	AuthenticateUser(login string, password string) (*auth_ldap.User, []string, error) // user config
}

type TokenGenerator interface {
	GenerateToken(user *auth_ldap.User, permissions []string) (string, error)
}

type TokenValidator interface {
	ValidateToken(tokenString string) (*auth_ldap.CustomClaims, error)
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func HandleLogin(log *slog.Logger, auth Authenticator, tokenGen TokenGenerator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handler.auth.HandleLogin"

		var req LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error": "Некорректный JSON"}`, http.StatusBadRequest)
			return
		}

		// 1. Аутентификация через LDAP
		user, groups, err := auth.AuthenticateUser(req.Username, req.Password)
		if err != nil {
			log.Warn(op, slog.String("user", req.Username), slog.String("error", err.Error()))
			http.Error(w, `{"error": "Неверный логин или пароль"}`, http.StatusUnauthorized)
			return
		}

		// 2. Маппинг групп в права (вызываем чистую функцию из пакета auth_ldap)
		permissions := auth_ldap.MapGroupsToPermission(groups)
		if len(permissions) == 0 {
			permissions = groups
		}

		// 3. Генерация токена
		token, err := tokenGen.GenerateToken(user, permissions)
		if err != nil {
			log.Error(op, slog.String("error", err.Error()))
			http.Error(w, `{"error": "Ошибка сервера при генерации токена"}`, http.StatusInternalServerError)
			return
		}

		// 4. Ответ
		resp := map[string]interface{}{
			"token": token,
			"user": map[string]interface{}{
				"uid":         user.UID,
				"full_name":   user.FullName,
				"groups":      groups,
				"permissions": permissions,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

// HandleTestProtected - тестовый защищенный маршрут
func HandleTestProtected(log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handler.auth.HandleTestProtected"

		// Достаем claims из контекста (их туда положил middleware)
		claims, ok := r.Context().Value("user_claims").(*auth_ldap.CustomClaims)
		if !ok {
			log.Error(op, slog.String("error", "claims not found in context"))
			http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
			return
		}

		resp := map[string]interface{}{
			"message":     "Success! Вы успешно аутентифицированы и авторизованы.",
			"user_uid":    claims.UID,
			"user_name":   claims.FullName,
			"permissions": claims.Permissions,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}
