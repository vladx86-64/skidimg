package handler

import (
	"net/http"

	"github.com/go-chi/chi"
)

var r *chi.Mux

func RegisterRoutes(handler *handler) *chi.Mux {
	r = chi.NewMux()
	tokenMaker := handler.TokenMaker

	r.Route("/users", func(r chi.Router) {
		r.Post("/", handler.createUser)
		r.Post("/login", handler.loginUser)

		r.Group(func(r chi.Router) {
			r.Use(GetAdminMiddlewareFunc(tokenMaker))
			r.Get("/", handler.listUsers)
			r.Route("/{id}", func(r chi.Router) {
				r.Delete("/", handler.deleteUser)
			})
		})

		r.Group(func(r chi.Router) {
			r.Use(GetAuthMiddlewareFUnc(tokenMaker))
			r.Patch("/", handler.updateUser)
			r.Post("/logout", handler.logoutUser)
		})
	})

	r.Group(func(r chi.Router) {
		r.Use(GetAuthMiddlewareFUnc(tokenMaker))
		r.Route("/tokens", func(r chi.Router) {
			r.Post("/renew", handler.renewAccessToken)
			r.Post("/revoke", handler.revokeSession)
		})
	})

	r.Group(func(r chi.Router) {
		r.Use(InjectOptionalClaims(tokenMaker))
		r.Use(InjectLayoutTemplateData())
		r.Get("/upload", handler.renderUploadPage)
		r.Post("/upload", handler.uploadImage)
	})

	r.Handle("/uploads/*", http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads"))))

	r.Group(func(r chi.Router) {
		r.Use(InjectOptionalClaims(tokenMaker)) // чтобы не было 401 для гостей
		r.Use(InjectLayoutTemplateData())
		r.Get("/gallery/{id}", handler.viewImage)
	})

	r.Post("/login", handler.handleLogin)
	r.Post("/register", handler.handleRegister)

	r.Group(func(r chi.Router) {

		r.Use(GetAuthMiddlewareFUnc(tokenMaker))
		r.Use(InjectLayoutTemplateData())
		r.Get("/gallery", handler.handleGalleryPage)
	})

	r.Group(func(r chi.Router) {

		r.Use(InjectLayoutTemplateData())
		r.Get("/register", handler.RenderRegisterPage)
		r.Get("/login", handler.RenderLoginPage)
	})

	r.Group(func(r chi.Router) {
		r.Use(GetAuthMiddlewareFUnc(tokenMaker))
		r.Use(InjectLayoutTemplateData()) // ← добавь ЭТО
		r.Get("/profile", handler.renderProfilePage)
	})

	return r
}

func Start(address string) error {
	return http.ListenAndServe(address, r)
}
