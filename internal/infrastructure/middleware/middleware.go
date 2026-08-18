package middleware

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/M1ralai/notly-api/internal/common/utils"
	"github.com/golang-jwt/jwt/v5"
)

var (
	jwtKey  []byte
	jwtOnce sync.Once
)

func getJWTKey() []byte {
	jwtOnce.Do(func() {
		key := os.Getenv("JWT_SECRET")
		if key == "" {
			log.Fatal("JWT_SECRET environment variable is not set")
		}
		jwtKey = []byte(key)
	})
	return jwtKey
}

type Claims struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isPublicAuthPath(r.URL.Path) || r.URL.Path == "/health" || r.Method == "OPTIONS" {
			next.ServeHTTP(w, r)
			return
		}

		authHeader := r.Header.Get("Authorization")

		if authHeader == "" {
			resp := utils.ErrorResponse("UNAUTHORIZED", "Giriş yapmanız gerekiyor", "Token eksik")
			utils.Return(w, http.StatusUnauthorized, resp)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			resp := utils.ErrorResponse("UNAUTHORIZED", "Geçersiz token formatı", "Bearer token bekleniyor")
			utils.Return(w, http.StatusUnauthorized, resp)
			return
		}
		tokenString := parts[1]

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return getJWTKey(), nil
		}, jwt.WithLeeway(5*time.Minute))

		if err != nil || !token.Valid {
			resp := utils.ErrorResponse("UNAUTHORIZED", "Oturum süresi dolmuş", "Token geçersiz veya süresi dolmuş")
			utils.Return(w, http.StatusUnauthorized, resp)
			return
		}

		ctx := context.WithValue(r.Context(), utils.RoleKey, claims.Role)
		ctx = context.WithValue(ctx, utils.UsernameKey, claims.Username)
		ctx = context.WithValue(ctx, utils.UserIDKey, claims.UserID)

		// Extract Connection ID for WebSocket echo prevention
		connectionID := r.Header.Get("X-Connection-ID")
		if connectionID != "" {
			ctx = context.WithValue(ctx, utils.ConnectionIDKey, connectionID)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func isPublicAuthPath(path string) bool {
	switch path {
	case "/api/auth/login",
		"/api/auth/register",
		"/api/auth/verify",
		"/api/auth/resend-code",
		"/api/auth/refresh",
		"/api/auth/logout",
		"/auth/login",
		"/auth/register":
		return true
	default:
		return false
	}
}

func TimeoutMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip timeout for WebSocket - WebSocket connections are long-lived
		if r.URL.Path == "/ws" {
			next.ServeHTTP(w, r)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
