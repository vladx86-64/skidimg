package handler

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"skidimg/internal/token"
	"strings"
	"time"
)

type authKey struct{}
type layoutKey struct{}

type LayoutTemplateData struct {
	IsAuthenticated bool
	Username        string
	Content         any
	Title           string
}

func GetAuthMiddlewareFUnc(tokenMaker *token.JWTMaker) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			claims, err := verifyClaimFromAutheader(r, tokenMaker)
			if err != nil {
				http.Error(w, fmt.Sprintf("error verifying token %v", err), http.StatusUnauthorized)
				return
			}

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

				ctx := context.WithValue(r.Context(), authKey{}, claims)
				r = r.WithContext(ctx)
			}

			next.ServeHTTP(w, r)
		})
	}
}

func GetAdminMiddlewareFunc(tokenMaker *token.JWTMaker) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			claims, err := verifyClaimFromAutheader(r, tokenMaker)
			if err != nil {
				http.Error(w, fmt.Sprintf("error verifying token %v", err), http.StatusUnauthorized)
				return
			}

			if !claims.IsAdmin {
				http.Error(w, "user is not an admin", http.StatusForbidden)
				return
			}

			ctx := context.WithValue(r.Context(), authKey{}, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func verifyClaimFromAutheader(r *http.Request, tokenMaker *token.JWTMaker) (*token.UserClaims, error) {
	authHeader := r.Header.Get("Authorization")

	if authHeader != "" {
		fields := strings.Fields(authHeader)
		if len(fields) == 2 && fields[0] == "Bearer" {
			return tokenMaker.VerifyToken(fields[1])
		}
	}

	cookie, err := r.Cookie("access_token")
	if err == nil && cookie.Value != "" {
		return tokenMaker.VerifyToken(cookie.Value)
	}

	return nil, fmt.Errorf("missing token in header and cookie")
}

func InjectLayoutTemplateData() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var layoutData LayoutTemplateData

			if val := r.Context().Value(authKey{}); val != nil {
				layoutData.IsAuthenticated = true
			}

			ctx := context.WithValue(r.Context(), layoutKey{}, layoutData)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetAuthWithRefreshMiddleware — middleware, который:
//  1. Проверяет access_token.
//  2. Если он истёк (ErrTokenExpired), берёт refresh_token из cookie, верифицирует его,
//     создаёт новый access_token и проставляет его в cookie.
//  3. Кладёт claims в контекст и зовёт следующий handler.
func GetAuthWithRefreshMiddleware(tokenMaker *token.JWTMaker) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			raw := getBearerOrCookie(r, "access_token")

			var claims *token.UserClaims
			var err error

			if raw == "" {

				err = jwt.ErrTokenExpired
			} else {
				claims, err = tokenMaker.VerifyToken(raw)
			}

			if err != nil {
				// 1) expired или не было токена — рефрешим
				if errors.Is(err, jwt.ErrTokenExpired) {
					rc, err2 := r.Cookie("refresh_token")
					if err2 != nil {
						http.Error(w, "нет access_token и нет refresh_token", http.StatusUnauthorized)
						return
					}
					refreshClaims, err2 := tokenMaker.VerifyToken(rc.Value)
					if err2 != nil {
						http.Error(w, "invalid refresh token", http.StatusUnauthorized)
						return
					}
					// TODO: проверить сессию в БД, что она не отозвана

					newTok, newClaims, err2 := tokenMaker.CreateToken(
						refreshClaims.ID, refreshClaims.Email, refreshClaims.IsAdmin, time.Minute*15,
					)
					if err2 != nil {
						http.Error(w, "не удалось создать новый access token", http.StatusInternalServerError)
						return
					}

					http.SetCookie(w, &http.Cookie{
						Name:     "access_token",
						Value:    newTok,
						HttpOnly: true,
						Path:     "/",
						SameSite: http.SameSiteLaxMode,
						Expires:  newClaims.ExpiresAt.Time,
					})
					claims = newClaims
				} else {

					http.Error(w, fmt.Sprintf("error verifying token: %v", err), http.StatusUnauthorized)
					return
				}
			}

			ctx := context.WithValue(r.Context(), authKey{}, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func getBearerOrCookie(r *http.Request, cookieName string) string {
	if h := r.Header.Get("Authorization"); h != "" {
		parts := strings.Fields(h)
		if len(parts) == 2 && parts[0] == "Bearer" {
			return parts[1]
		}
	}
	if c, err := r.Cookie(cookieName); err == nil {
		return c.Value
	}
	return ""
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
