package api

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (api *Api) BindRoutes() {
	api.Router.Use(middleware.RequestID, middleware.Recoverer, middleware.Logger)

	// By default csrfMiddleware will look for a header containing the
	// X-CSRF-Token with the csrf token inside of it.
	// csrfMiddleware := csrf.Protect(
	// 	[]byte(os.Getenv("GOBID_CSRF_KEY")),
	// 	csrf.Secure(false),
	// )

	api.Router.Use(api.Session.LoadAndSave)
	// api.Router.Use(csrfMiddleware, api.Session.LoadAndSave)

	api.Router.With(api.AuthMiddleware).Get("/ws/subscribe/{product_id}", api.handleSubcribeUserToAuction)

	// /api/subscribe/10 -> Guitarra ibanez pika
	api.Router.Route("/api", func(r chi.Router) {
		r.Get("/csrf-token", api.handleGetCSRFToken)
		r.Route("/v1", func(r chi.Router) {
			r.Route("/users", func(r chi.Router) {
				r.Post("/signup", api.handleSignUpUser)
				r.Post("/login", api.handleLoginUser)

				// the user needs to be logged in.
				r.With(api.AuthMiddleware).Post("/logout", api.handleLogOut)
			})

			r.Route("/products", func(r chi.Router) {
				r.Group(func(r chi.Router) {
					r.Get("/list", api.handleListProducts)
					r.Get("/{id}", api.handleListProductById)
				})

				r.Group(func(r chi.Router) {
					r.Use(api.AuthMiddleware)
					r.Post("/", api.handleCreateProduct)
				})
			})
		})
	})
}
