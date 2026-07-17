package auth

import (
	"context"
	"net/http"
	"strings"

	"garuda-hacks/backend/users"
)

type contextKey string

const claimsContextKey contextKey = "auth_claims"

func Authenticate(tokenManager *TokenManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenString, err := bearerToken(r.Header.Get("Authorization"))
			if err != nil {
				writeError(w, http.StatusUnauthorized, err.Error())
				return
			}

			claims, err := tokenManager.ValidateAccessToken(tokenString)
			if err != nil {
				writeError(w, http.StatusUnauthorized, ErrInvalidToken.Error())
				return
			}

			ctx := context.WithValue(r.Context(), claimsContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func Authorize(roles ...users.UserRole) func(http.Handler) http.Handler {
	return RequireRole(roles...)
}

func RequireRole(roles ...users.UserRole) func(http.Handler) http.Handler {
	allowedRoles := make(map[users.UserRole]struct{}, len(roles))
	for _, role := range roles {
		normalized := users.UserRole(strings.ToUpper(strings.TrimSpace(string(role))))
		allowedRoles[normalized] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := ClaimsFromContext(r.Context())
			if !ok {
				writeError(w, http.StatusUnauthorized, ErrMissingToken.Error())
				return
			}

			if len(allowedRoles) == 0 {
				next.ServeHTTP(w, r)
				return
			}

			if _, ok := allowedRoles[claims.Role]; !ok {
				writeError(w, http.StatusForbidden, ErrForbidden.Error())
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func RequirePermission(permissions ...string) func(http.Handler) http.Handler {
	required := make(map[string]struct{}, len(permissions))
	for _, p := range permissions {
		required[strings.ToLower(strings.TrimSpace(p))] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := ClaimsFromContext(r.Context())
			if !ok {
				writeError(w, http.StatusUnauthorized, ErrMissingToken.Error())
				return
			}

			userPerms := GetPermissionsForRole(claims.Role)
			userPermsMap := make(map[string]struct{}, len(userPerms))
			for _, p := range userPerms {
				userPermsMap[p] = struct{}{}
			}

			hasAll := true
			for p := range required {
				if _, exists := userPermsMap[p]; !exists {
					hasAll = false
					break
				}
			}

			if !hasAll {
				writeError(w, http.StatusForbidden, ErrForbidden.Error())
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func GetCurrentUser(ctx context.Context) (*Claims, bool) {
	return ClaimsFromContext(ctx)
}

func ClaimsFromContext(ctx context.Context) (*Claims, bool) {
	claims, ok := ctx.Value(claimsContextKey).(*Claims)
	return claims, ok
}

func GetPermissionsForRole(role users.UserRole) []string {
	switch role {
	case users.RoleAdmin:
		return []string{
			"profile:read", "profile:write", "profile:delete",
			"org:create", "org:read", "org:write", "org:manage", "org:invite", "org:kick",
			"nib:submit", "nib:verify", "users:list", "users:read",
		}
	case users.RoleBuyer:
		return []string{
			"profile:read", "profile:write",
			"org:create", "org:read", "org:write",
			"nib:submit",
		}
	case users.RoleProducer, users.RoleFarmer:
		return []string{
			"profile:read", "profile:write",
			"org:create", "org:read", "org:write",
			"nib:submit",
		}
	default:
		return []string{"profile:read"}
	}
}

func bearerToken(authorizationHeader string) (string, error) {
	const prefix = "Bearer "
	if !strings.HasPrefix(authorizationHeader, prefix) {
		return "", ErrMissingToken
	}

	token := strings.TrimSpace(strings.TrimPrefix(authorizationHeader, prefix))
	if token == "" {
		return "", ErrMissingToken
	}

	return token, nil
}
