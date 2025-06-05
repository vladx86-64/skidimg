package handler

import (
	"context"
	"fmt"
	"net/http"
	"skidimg/internal/token"
	"strings"
)

type authKey struct{}

func GetAuthMiddlewareFUnc(tokenMaker *token.JWTMaker) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			// read authroization header
			// verify the oken
			claims, err := verifyClaimFromAutheader(r, tokenMaker)
			if err != nil {
				http.Error(w, fmt.Sprintf("error verifying token %v", err), http.StatusUnauthorized)
				return
			}

			// pass the paylaod/claim down the context
			ctx := context.WithValue(r.Context(), authKey{}, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func InjectOptionalClaims(tokenMaker *token.JWTMaker) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, err := verifyClaimFromAutheader(r, tokenMaker)
			if err == nil && claims != nil {
				// если токен валидный — кладём в контекст
				ctx := context.WithValue(r.Context(), authKey{}, claims)
				r = r.WithContext(ctx)
			}
			// если токена нет — просто продолжаем без claims
			next.ServeHTTP(w, r)
		})
	}
}

func GetAdminMiddlewareFunc(tokenMaker *token.JWTMaker) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			// read authroization header
			// verify the oken
			claims, err := verifyClaimFromAutheader(r, tokenMaker)
			if err != nil {
				http.Error(w, fmt.Sprintf("error verifying token %v", err), http.StatusUnauthorized)
				return
			}

			if !claims.IsAdmin {
				http.Error(w, "user is not an admin", http.StatusForbidden)
				return
			}

			// pass the paylaod/claim down the context
			ctx := context.WithValue(r.Context(), authKey{}, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// func verifyClaimFromAutheader(r *http.Request, tokenMaker *token.JWTMaker) (*token.UserClaims, error) {
// 	authHeader := r.Header.Get("Authorization")
// 	if authHeader == "" {
// 		return nil, fmt.Errorf("auth header missing")
// 	}

// 	fields := strings.Fields(authHeader)
// 	if len(fields) != 2 || fields[0] != "Bearer" {
// 		return nil, fmt.Errorf("invalis auth header")
// 	}

// 	token := fields[1]
// 	claims, err := tokenMaker.VerifyToken(token)
// 	if err != nil {
// 		return nil, fmt.Errorf("invalid token %w", err)
// 	}

// 	return claims, nil
// }

func verifyClaimFromAutheader(r *http.Request, tokenMaker *token.JWTMaker) (*token.UserClaims, error) {
	authHeader := r.Header.Get("Authorization")

	// ✅ Если есть Bearer, используем его
	if authHeader != "" {
		fields := strings.Fields(authHeader)
		if len(fields) == 2 && fields[0] == "Bearer" {
			return tokenMaker.VerifyToken(fields[1])
		}
	}

	// ✅ Иначе — пытаемся вытащить access_token из cookie
	cookie, err := r.Cookie("access_token")
	if err == nil && cookie.Value != "" {
		return tokenMaker.VerifyToken(cookie.Value)
	}

	return nil, fmt.Errorf("missing token in header and cookie")
}
