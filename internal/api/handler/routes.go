package handler

import (
	"net/http"

	"github.com/go-chi/chi"
)

var r *chi.Mux

// func RegisterRoutes(handler *handler) *chi.Mux {
// 	r = chi.NewMux()
// 	tokenMaker := handler.TokenMaker
//
// 	r.Route("/users", func(r chi.Router) {
// 		r.Post("/", handler.createUser)
// 		r.Post("/login", handler.loginUser)
//
// 		r.Group(func(r chi.Router) {
// 			r.Use(GetAdminMiddlewareFunc(tokenMaker))
// 			r.Get("/", handler.listUsers)
// 			r.Route("/{id}", func(r chi.Router) {
// 				r.Delete("/", handler.deleteUser)
// 			})
// 		})
//
// 		r.Group(func(r chi.Router) {
// 			r.Use(GetAuthMiddlewareFUnc(tokenMaker))
// 			r.Patch("/", handler.updateUser)
// 			r.Post("/logout", handler.logoutUser)
// 		})
// 	})
//
// 	r.Group(func(r chi.Router) {
// 		r.Use(GetAuthMiddlewareFUnc(tokenMaker))
// 		r.Route("/tokens", func(r chi.Router) {
// 			r.Post("/renew", handler.renewAccessToken)
// 			r.Post("/revoke", handler.revokeSession)
// 		})
// 	})
//
// 	r.Group(func(r chi.Router) {
// 		r.Use(InjectOptionalClaims(tokenMaker))
// 		r.Use(InjectLayoutTemplateData())
// 		r.Get("/upload", handler.renderUploadPage)
// 		r.Post("/upload", handler.uploadImage)
// 	})
//
// 	r.Handle("/uploads/*", http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads"))))
//
// 	r.Group(func(r chi.Router) {
// 		r.Use(InjectOptionalClaims(tokenMaker)) // чтобы не было 401 для гостей
// 		r.Use(InjectLayoutTemplateData())
// 		r.Get("/gallery/{id}", handler.viewImage)
// 	})
//
// 	r.Post("/login", handler.handleLogin)
// 	r.Post("/register", handler.handleRegister)
//
// 	r.Group(func(r chi.Router) {
//
// 		r.Use(GetAuthMiddlewareFUnc(tokenMaker))
// 		r.Use(InjectLayoutTemplateData())
// 		r.Get("/gallery", handler.handleGalleryPage)
// 	})
//
// 	r.Group(func(r chi.Router) {
//
// 		r.Use(InjectLayoutTemplateData())
// 		r.Get("/register", handler.RenderRegisterPage)
// 		r.Get("/login", handler.RenderLoginPage)
// 	})
//
// 	r.Group(func(r chi.Router) {
// 		r.Use(GetAuthMiddlewareFUnc(tokenMaker))
// 		r.Use(InjectLayoutTemplateData())
// 		r.Get("/profile", handler.renderProfilePage)
// 	})
//
// 	// r.Group(func(r chi.Router) {
// 	// 	r.Use(GetAuthMiddlewareFUnc(tokenMaker))
// 	// 	r.Use(InjectLayoutTemplateData())
// 	// 	r.Get("/album", handler.renderAlbumPage)
// 	// })
//
// 	r.Route("/albums", func(r chi.Router) {
// 		r.Use(GetAuthMiddlewareFUnc(tokenMaker))
// 		r.Use(InjectLayoutTemplateData())
//
// 		// /albums
// 		r.Get("/", handler.renderUserAlbums)   // список
// 		r.Post("/", handler.handleCreateAlbum) // создание
//
// 		// /albums/{id}
// 		r.Route("/{id}", func(r chi.Router) {
// 			r.Get("/", handler.viewAlbum)            // просмотр альбома
// 			r.Post("/add", handler.handleAddToAlbum) // добавление изображений
//
// 			r.Post("/remove/{image_id}", handler.handleRemoveFromAlbum)
// 			r.Post("/delete", handler.handleDeleteAlbum)
// 		})
// 	})
//
// 	return r
// }

func RegisterRoutes(h *handler) *chi.Mux {
	r = chi.NewMux()
	tm := h.TokenMaker

	// Пользователи
	r.Route("/users", func(r chi.Router) {
		r.Post("/", h.createUser)
		r.Post("/login", h.loginUser) // JSON-логин

		r.Group(func(r chi.Router) {
			r.Use(GetAuthWithRefreshMiddleware(tm))
			r.Patch("/", h.updateUser)
			r.Post("/logout", h.logoutUser)
		})

		r.Group(func(r chi.Router) {
			r.Use(GetAdminMiddlewareFunc(tm))
			r.Get("/", h.listUsers)
			r.Delete("/{id}", h.deleteUser)
		})
	})

	// Токены (рефреш, отзыв)
	r.Group(func(r chi.Router) {
		r.Use(GetAuthWithRefreshMiddleware(tm))
		r.Post("/tokens/renew", h.renewAccessToken)
		r.Post("/tokens/revoke", h.revokeSession)
	})

	// Страницы и загрузки
	r.Group(func(r chi.Router) {
		r.Use(InjectOptionalClaims(tm))
		r.Use(InjectLayoutTemplateData())
		r.Get("/upload", h.renderUploadPage)
		r.Post("/upload", h.uploadImage)
	})

	r.Group(func(r chi.Router) {
		r.Use(InjectOptionalClaims(tm))
		r.Use(InjectLayoutTemplateData())
		r.Get("/faq", h.renderFAQPage)
		r.Get("/terms", h.renderTermsPage)
	})

	r.Handle("/uploads/*", http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads"))))

	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("./web/static"))))

	// Галерея: гостям и залогиненным разный вид
	r.Group(func(r chi.Router) {
		r.Use(InjectOptionalClaims(tm))
		r.Use(InjectLayoutTemplateData())
		r.Get("/gallery/{id}", h.viewImage)
	})

	// HTML-страницы регистрации/логина
	r.Group(func(r chi.Router) {
		r.Use(InjectLayoutTemplateData())
		r.Get("/register", h.RenderRegisterPage)
		r.Get("/login", h.RenderLoginPage)
		r.Post("/register", h.handleRegister)
		r.Post("/login", h.handleLogin) // form-логин
	})

	// Защищённая галерея, профиль, альбомы
	r.Group(func(r chi.Router) {
		r.Use(GetAuthWithRefreshMiddleware(tm))
		r.Use(InjectLayoutTemplateData())

		r.Get("/gallery", h.handleGalleryPage)
		r.Get("/profile", h.renderProfilePage)

		r.Route("/albums", func(r chi.Router) {
			r.Get("/", h.renderUserAlbums)
			r.Post("/", h.handleCreateAlbum)
			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", h.viewAlbum)
				r.Post("/add", h.handleAddToAlbum)
				r.Post("/remove/{image_id}", h.handleRemoveFromAlbum)
				r.Post("/delete", h.handleDeleteAlbum)
			})
		})
	})

	return r
}
func Start(address string) error {
	return http.ListenAndServe(address, r)
}
