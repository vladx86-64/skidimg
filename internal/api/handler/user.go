package handler

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"skidimg/internal/model"
	"skidimg/internal/security"
	"skidimg/internal/token"
	"strconv"
	"time"

	"github.com/go-chi/chi"
)

// func toStorageUser(u model.UserReq) *model.User {
// 	return &model.User{
// 		Name:     u.Name,
// 		Email:    u.Email,
// 		Password: u.Password,
// 		IsAdmin:  u.IsAdmin,
// 	}
// }

func toTimePtr(t time.Time) *time.Time {
	return &t
}

func patchUserReq(user *model.User, u model.UserReq) {
	if u.Name != "" {
		user.Name = u.Name
	}
	if u.Email != "" {
		user.Email = u.Email
	}
	if u.Password != "" {
		hashedPassword, err := security.HashPassword(u.Password)
		if err != nil {
			panic(err)
		}
		user.Password = hashedPassword
	}
	if u.IsAdmin {
		user.IsAdmin = u.IsAdmin
	}
	user.UpdatedAt = toTimePtr(time.Now())
}

func (h *handler) createUser(w http.ResponseWriter, r *http.Request) {
	var u model.UserReq
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		http.Error(w, "Bad request ", http.StatusBadRequest)
		return
	}

	hashed, err := security.HashPassword(u.Password)
	if err != nil {
		http.Error(w, "Error hasing password ", http.StatusInternalServerError)
		return
	}
	u.Password = hashed

	createdUser, err := h.server.CreateUser(h.ctx, u.ToStorage())
	if err != nil {
		http.Error(w, fmt.Sprintf("Error creating user %v", err), http.StatusInternalServerError)
		return
	}

	res := createdUser.ToRes()
	w.Header().Set("Content-Type", "application/json") // Устанавливаем заголовок
	w.WriteHeader(http.StatusCreated)                  // отпарвляем статус
	json.NewEncoder(w).Encode(res)                     // сериализуем и пишем прямо в http ответ
}

func (h *handler) listUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.server.ListUsers(h.ctx)
	if err != nil {
		http.Error(w, "Error listing users", http.StatusInternalServerError)
		return
	}

	var res model.ListUserRes // response with list  of users

	for i := range users {
		res.Users = append(res.Users, *users[i].ToRes())
	}

	w.Header().Set("Content-Type", "application/json") // Устанавливаем заголовок
	json.NewEncoder(w).Encode(res)                     // сериализуем и пишем прямо в http ответ и по дефолту отпраялем статус 200
}

func (h *handler) updateUser(w http.ResponseWriter, r *http.Request) {
	// @TODO  получать email пользовател с payload то
	var u model.UserReq
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		http.Error(w, "Bad request ", http.StatusBadRequest)
		return
	}

	// теперь мы читаем claims
	claims := r.Context().Value(authKey{}).(*token.UserClaims)

	user, err := h.server.GetUser(h.ctx, claims.Email)
	if err != nil {
		http.Error(w, "Error creating user", http.StatusInternalServerError)
		return
	}

	patchUserReq(user, u)
	if user.Email == "" {
		user.Email = claims.Email
	}

	updatedUser, err := h.server.UpdateUser(h.ctx, user)
	if err != nil {
		http.Error(w, "Error updating user", http.StatusInternalServerError)
		return
	}

	res := updatedUser.ToRes()
	w.Header().Set("Content-Type", "application/json") // Устанавливаем заголовок
	w.WriteHeader(http.StatusOK)                       // отпарвляем статус
	json.NewEncoder(w).Encode(res)                     // сериализуем и пишем прямо в http ответ

}

func (h *handler) deleteUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	i, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		panic(err)
	}

	err = h.server.DeleteUser(h.ctx, i)
	if err != nil {
		http.Error(w, "Error deleting user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *handler) loginUser(w http.ResponseWriter, r *http.Request) {
	var u model.LoginUserReq
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		http.Error(w, "Bad request ", http.StatusBadRequest)
		return
	}

	gu, err := h.server.GetUser(h.ctx, u.Email)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error getting user data while logging in %v", err), http.StatusInternalServerError)
		return
	}

	err = security.CheckPassword(u.Password, gu.Password)
	if err != nil {
		http.Error(w, "Wrong passwrord", http.StatusUnauthorized)
		return
	}

	// создаем jWT токен и отпрвлеям как отвтет
	accessToken, accessClaims, err := h.TokenMaker.CreateToken(gu.ID, gu.Email, gu.IsAdmin, time.Minute*15)
	if err != nil {
		http.Error(w, "Error creating token", http.StatusInternalServerError)
		return
	}

	resfreshToken, refreshClaims, err := h.TokenMaker.CreateToken(gu.ID, gu.Email, gu.IsAdmin, time.Hour*24)
	if err != nil {
		http.Error(w, "Error creating token", http.StatusInternalServerError)
		return
	}

	session, err := h.server.CreateSession(h.ctx, &model.Session{
		ID:           refreshClaims.RegisteredClaims.ID,
		UserEmail:    gu.Email,
		RefreshToken: resfreshToken,
		IsRevoked:    false,
		ExpiresAt:    refreshClaims.ExpiresAt.Time,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("error creating session %v", err), http.StatusInternalServerError)
		return
	}

	res := model.LoginUserRes{
		SessionID:             session.ID,
		AccessToken:           accessToken,
		RefreshTOken:          resfreshToken,
		AccessTokenExpiresAt:  accessClaims.ExpiresAt.Time,
		RefreshTokenExpiresAt: refreshClaims.ExpiresAt.Time,
		User:                  *gu.ToRes(),
	}

	w.Header().Set("Content-Type", "application/json") // Устанавливаем заголовок
	w.WriteHeader(http.StatusOK)                       // отпарвляем статус
	json.NewEncoder(w).Encode(res)                     // сериализуем и пишем прямо в http ответ
}

func (h *handler) logoutUser(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(authKey{}).(*token.UserClaims)

	err := h.server.DeleteSession(h.ctx, claims.RegisteredClaims.ID)
	if err != nil {
		http.Error(w, "Error deleting session", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *handler) renewAccessToken(w http.ResponseWriter, r *http.Request) {
	var req model.RenewAccessTokenReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request ", http.StatusBadRequest)
		return
	}

	refreshClaims, err := h.TokenMaker.VerifyToken(req.RefreshToken)
	if err != nil {
		http.Error(w, "Error verifying token", http.StatusUnauthorized)
		return
	}

	session, err := h.server.GetSession(h.ctx, refreshClaims.RegisteredClaims.ID)
	if err != nil {
		http.Error(w, "Error getting session", http.StatusInternalServerError)
		return
	}

	if session.IsRevoked {
		http.Error(w, "Session revoked", http.StatusUnauthorized)
		return
	}

	if session.UserEmail != refreshClaims.Email {
		http.Error(w, "invalid session", http.StatusUnauthorized)
		return
	}

	accessToken, accessClaims, err := h.TokenMaker.CreateToken(refreshClaims.ID, refreshClaims.Email, refreshClaims.IsAdmin, time.Minute*15)
	if err != nil {
		http.Error(w, "error creating token", http.StatusInternalServerError)
		return
	}

	res := model.RenewAccessTokenRes{
		AccessToken:          accessToken,
		AccessTokenExpiresAt: accessClaims.ExpiresAt.Time,
	}

	w.Header().Set("Content-Type", "application/json") // Устанавливаем заголовок
	w.WriteHeader(http.StatusOK)                       // отпарвляем статус
	json.NewEncoder(w).Encode(res)
}

func (h *handler) revokeSession(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(authKey{}).(*token.UserClaims)

	err := h.server.DeleteSession(h.ctx, claims.RegisteredClaims.ID)
	if err != nil {
		http.Error(w, "error deleting session", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *handler) RenderLoginPage(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles(
		"web/templates/layout.html",
		"web/templates/login.html",
	)
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
	tmpl.ExecuteTemplate(w, "layout", nil)
}

func (h *handler) RenderRegisterPage(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles(
		"web/templates/layout.html",
		"web/templates/register.html",
	)
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
	tmpl.ExecuteTemplate(w, "layout", nil)
}
