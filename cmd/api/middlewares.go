package main

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"social/internal/store"
	"strconv"
	"strings"

	"github.com/go-chi/cors"
	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrorMissingAuthHeader   = errors.New("missing authorization header")
	ErrorMalformedAuthHeader = errors.New("authorization header is malformed")
	ErrorInvalidCredentials  = errors.New("invalid credentials")
	ErrorBearerTokenMissing  = errors.New("missing bearer token from header")
)

type userCtx string

var (
	userCtxKey userCtx = "userCtx"
)

func corsMiddleware(h http.Handler) func(http.Handler) http.Handler {
	return cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	})
}

func (app *application) basicAuthMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				app.statusBasicUnauthorized(w, r, ErrorMissingAuthHeader)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Basic" {
				app.statusBasicUnauthorized(w, r, ErrorMalformedAuthHeader)
				return
			}

			decoded, err := base64.StdEncoding.DecodeString(parts[1])
			if err != nil {
				app.statusBasicUnauthorized(w, r, err)
				return
			}

			username := app.config.auth.basic.user
			passwd := app.config.auth.basic.password

			creds := strings.SplitN(string(decoded), ":", 2)
			if len(creds) != 2 || creds[0] != username || creds[1] != passwd {
				app.statusBasicUnauthorized(w, r, ErrorInvalidCredentials)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func (app *application) AuthTokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			app.statusUnauthorized(w, r, ErrorMissingAuthHeader)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 {
			app.statusUnauthorized(w, r, ErrorMalformedAuthHeader)
			return
		}

		if parts[0] != "Bearer" {
			app.statusUnauthorized(w, r, ErrorBearerTokenMissing)
			return
		}

		token := parts[1]
		jwtToken, err := app.authenticator.ValidateToken(token)
		if err != nil {
			app.statusUnauthorized(w, r, err)
			return
		}

		claims := jwtToken.Claims.(jwt.MapClaims)

		userID, err := strconv.ParseInt(fmt.Sprintf("%.f", claims["sub"]), 10, 64)
		if err != nil {
			app.statusUnauthorized(w, r, err)
			return
		}

		ctx := r.Context()

		user, err := app.getUser(ctx, userID)
		if err != nil {
			app.statusUnauthorized(w, r, err)
			return
		}

		ctx = context.WithValue(ctx, userCtxKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (app *application) checkPostOwnership(requiredRole string, next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user := ctx.Value(userCtxKey).(store.User)
		post := getPostFromCtx(r)

		if post.UserID == user.ID {
			next.ServeHTTP(w, r)
			return
		}

		allowed, err := app.checkRolePrecedence(ctx, &user, requiredRole)
		if err != nil {
			app.statusInternalServerError(w, r, err)
			return
		}

		if !allowed {
			app.forbiddenResponse(w, r, ErrorInvalidCredentials)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (app *application) checkRolePrecedence(ctx context.Context, user *store.User, requiredRole string) (bool, error) {
	role, err := app.store.Role.GetByName(ctx, requiredRole)
	if err != nil {
		return false, err
	}

	return user.Role.Level >= role.Level, nil
}

func (app *application) getUser(ctx context.Context, userID int64) (store.User, error) {
	if !app.config.redisCfg.enabled {
		return app.store.User.GetById(ctx, userID)
	}

	user, err := app.cacheStorage.Users.Get(ctx, userID)
	if err != nil && err != redis.Nil {
		return store.User{}, err
	}

	if user == nil {
		persistedUser, err := app.store.User.GetById(ctx, userID)
		if err != nil {
			return store.User{}, err
		}

		user = &persistedUser

		if err := app.cacheStorage.Users.Set(ctx, user); err != nil {
			return store.User{}, err
		}
	}

	return *user, nil
}

func (app *application) RateLimiterMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if app.config.rateLimiter.Enabled {
			if allow, retryAfter := app.rateLimiter.Allow(r.RemoteAddr); !allow {
				fmt.Println("entrou aqui")
				app.rateLimitExceededResponse(w, r, retryAfter.String())
			}
		}

		next.ServeHTTP(w, r)
	})
}
