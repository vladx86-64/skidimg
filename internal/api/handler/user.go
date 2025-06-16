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

	resfreshToken, refreshClaims, err := h.TokenMaker.CreateToken(gu.ID, gu.Email, gu.IsAdmin, time.Hour*24*30)
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

// func (h *handler) renewAccessToken(w http.ResponseWriter, r *http.Request) {
// 	cookie, err := r.Cookie("refresh_token")
// 	if err != nil || cookie.Value == "" {
// 		http.Error(w, "Missing refresh token", http.StatusUnauthorized)
// 		return
// 	}
//
// 	refreshClaims, err := h.TokenMaker.VerifyToken(cookie.Value)
// 	if err != nil {
// 		http.Error(w, "Invalid refresh token", http.StatusUnauthorized)
// 		return
// 	}
//
// 	session, err := h.server.GetSession(h.ctx, refreshClaims.RegisteredClaims.ID)
// 	if err != nil {
// 		http.Error(w, "Error getting session", http.StatusInternalServerError)
// 		return
// 	}
//
// 	if session.IsRevoked {
// 		http.Error(w, "Session revoked", http.StatusUnauthorized)
// 		return
// 	}
//
// 	if session.UserEmail != refreshClaims.Email {
// 		http.Error(w, "invalid session", http.StatusUnauthorized)
// 		return
// 	}
//
// 	accessToken, accessClaims, err := h.TokenMaker.CreateToken(refreshClaims.ID, refreshClaims.Email, refreshClaims.IsAdmin, time.Minute*15)
// 	if err != nil {
// 		http.Error(w, "error creating token", http.StatusInternalServerError)
// 		return
// 	}
//
// 	res := model.RenewAccessTokenRes{
// 		AccessToken:          accessToken,
// 		AccessTokenExpiresAt: accessClaims.ExpiresAt.Time,
// 	}
//
// 	w.Header().Set("Content-Type", "application/json") // Устанавливаем заголовок
// 	w.WriteHeader(http.StatusOK)                       // отпарвляем статус
// 	json.NewEncoder(w).Encode(res)
// }

func (h *handler) renewAccessToken(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil || cookie.Value == "" {
		http.Error(w, "Missing refresh token", http.StatusUnauthorized)
		return
	}

	refreshClaims, err := h.TokenMaker.VerifyToken(cookie.Value)
	if err != nil {
		http.Error(w, "Invalid refresh token", http.StatusUnauthorized)
		return
	}

	session, err := h.server.GetSession(h.ctx, refreshClaims.RegisteredClaims.ID)
	if err != nil {
		http.Error(w, "Error getting session", http.StatusInternalServerError)
		return
	}

	if session.IsRevoked || session.UserEmail != refreshClaims.Email {
		http.Error(w, "Invalid session", http.StatusUnauthorized)
		return
	}

	accessToken, accessClaims, err := h.TokenMaker.CreateToken(refreshClaims.ID, refreshClaims.Email, refreshClaims.IsAdmin, time.Minute*15)
	if err != nil {
		http.Error(w, "Error creating token", http.StatusInternalServerError)
		return
	}

	// 🎯 ВАЖНО: обновляем куку
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
		Expires:  accessClaims.ExpiresAt.Time,
	})

	// Можно ничего не возвращать — только статус
	w.WriteHeader(http.StatusNoContent)
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
	layoutData := r.Context().Value(layoutKey{}).(LayoutTemplateData)
	layoutData.Title = "SkidIMG - Login"

	tmpl.ExecuteTemplate(w, "layout", layoutData)
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

	layoutData := r.Context().Value(layoutKey{}).(LayoutTemplateData)
	layoutData.Title = "SkidIMG - Register"

	tmpl.ExecuteTemplate(w, "layout", layoutData)
}

func (h *handler) handleRegister(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad form data", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	email := r.FormValue("email")
	password := r.FormValue("password")

	if name == "" || email == "" || password == "" {
		http.Error(w, "Missing fields", http.StatusBadRequest)
		return
	}

	hashed, err := security.HashPassword(password)
	if err != nil {
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}

	u := model.UserReq{
		Name:     name,
		Email:    email,
		Password: hashed,
		IsAdmin:  false,
	}

	_, err = h.server.CreateUser(h.ctx, u.ToStorage())
	if err != nil {
		http.Error(w, fmt.Sprintf("Error creating user: %v", err), http.StatusInternalServerError)
		return
	}

	// 🎯 Просто редиректим на страницу логина
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (h *handler) handleLogin(w http.ResponseWriter, r *http.Request) {

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad form data", http.StatusBadRequest)
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")

	if email == "" || password == "" {
		http.Error(w, "Missing fields", http.StatusBadRequest)
		return
	}

	// Получаем пользователя
	gu, err := h.server.GetUser(h.ctx, email)
	if err != nil {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	// Проверяем пароль
	if err := security.CheckPassword(password, gu.Password); err != nil {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	// Генерируем токены
	accessToken, accessClaims, err := h.TokenMaker.CreateToken(gu.ID, gu.Email, gu.IsAdmin, time.Minute*15)
	if err != nil {
		http.Error(w, "Error creating token", http.StatusInternalServerError)
		return
	}

	refreshToken, refreshClaims, err := h.TokenMaker.CreateToken(gu.ID, gu.Email, gu.IsAdmin, time.Hour*24*30)
	if err != nil {
		http.Error(w, "Error creating refresh token", http.StatusInternalServerError)
		return
	}

	// Сохраняем сессию
	// ❗ используем строковый ID из RegisteredClaims
	_, err = h.server.CreateSession(h.ctx, &model.Session{
		ID:           refreshClaims.RegisteredClaims.ID, // ✅ string
		UserEmail:    gu.Email,
		RefreshToken: refreshToken,
		IsRevoked:    false,
		ExpiresAt:    refreshClaims.ExpiresAt.Time,
	})

	if err != nil {
		http.Error(w, "Error creating session", http.StatusInternalServerError)
		return
	}

	// Ставим access токен в cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		HttpOnly: true,
		// Secure:   true,                 // если на HTTPS, иначе можно убрать на локалке
		Path:     "/",                  // доступно для всех роутов
		SameSite: http.SameSiteLaxMode, // или Strict
		Expires:  accessClaims.ExpiresAt.Time,
	})

	// По желанию — refresh токен в отдельную cookie

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		HttpOnly: true,
		Path:     "/",                  // ✅ именно сюда ты стучишься
		SameSite: http.SameSiteLaxMode, // ✅ так тоже норм на http
		// Secure: true,                    // ❌ отключи на локалке
		Expires: refreshClaims.ExpiresAt.Time,
	})

	// http.Redirect(w, r, "/gallery", http.StatusSeeOther)
	// w.WriteHeader(http.StatusOK)

	http.Redirect(w, r, "/gallery", http.StatusSeeOther)
}

func (h *handler) renderTermsPage(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("web/templates/layout.html", "web/templates/terms.html")
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
	layoutData := r.Context().Value(layoutKey{}).(LayoutTemplateData)
	layoutData.Title = "SkidIMG - T&C"

	tmpl.ExecuteTemplate(w, "layout", layoutData)
}
func (h *handler) renderFAQPage(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("web/templates/layout.html", "web/templates/faq.html")
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
	layoutData := r.Context().Value(layoutKey{}).(LayoutTemplateData)

	layoutData.Title = "SkidIMG - FAQ"

	tmpl.ExecuteTemplate(w, "layout", layoutData)
}
