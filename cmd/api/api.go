package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/ozarabal/goSocial/docs"
	"github.com/ozarabal/goSocial/internal/mailer"
	"github.com/ozarabal/goSocial/internal/store"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"
)

type application struct {
	config 	config
	store 	store.Storage
	logger	*zap.SugaredLogger
	mailer	mailer.Client
}

type config struct {
	addr string
	apiURL string
	db dbConfig
	env string
	mail mailConfig
	frontendURL string
}

type mailConfig struct{
	sendGrid	sendGridConfig
	exp time.Duration
	fromEmail	string 
}

type sendGridConfig struct{
	apikey 		string
}

type dbConfig struct{
	addr 			string
	maxOpenConns 	int
	maxIdleConns 	int
	maxIdleTime 	string
}

func (app *application) mount() http.Handler {
	r := chi.NewRouter()
	
	// log
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// set a timeout value on middleware on the request context

	r.Use(middleware.Timeout(60 * time.Second))

	r.Route("/v1", func(r chi.Router){
		r.Get("/health", app.healthCheckHandler)
		docsURL := fmt.Sprintf("%s/swagger/doc.json", app.config.addr)
		r.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL(docsURL)))

		r.Route("/posts", func(r chi.Router){
			r.Post("/", app.createPostHandler)
			
			r.Route("/{postID}", func(r chi.Router){
				r.Use(app.postContextMiddleware)
					
				r.Get("/", app.getPostHandler)
				r.Delete("/", app.deletePostHandler)
				r.Patch("/", app.updatePostHandler)
			})
		})

		r.Route("/users", func(r chi.Router){
			r.Put("/activate/{token}", app.activateUserHandler)

			r.Route("/{userID}", func(r chi.Router){
				r.Use(app.userContextMiddleware)
				r.Get("/", app.getUserHandler)
				
				r.Put("/follow", app.followUserHandler)
				r.Put("/unfollow", app.unfollowUserHandler)
			})

			r.Group(func(r chi.Router) {
				
				r.Get("/feed", app.getUserFeedHandler)
			})
		})
		// public routes
		r.Route("/authentication", func(r chi.Router){
			r.Post("/user", app.registerUserHandler)
		})
	})	

	return r
}

func (app *application) run(mux http.Handler)error {
	// Docs
	docs.SwaggerInfo.Version = version
	docs.SwaggerInfo.Host = app.config.apiURL
	docs.SwaggerInfo.BasePath = "/v1"

	srv := &http.Server{
		Addr : app.config.addr,
		Handler: mux,
		ReadTimeout: time.Second * 10,
		WriteTimeout: time.Second * 30,
		IdleTimeout: time.Minute,
	}

	app.logger.Infow("server has started", "addr", app.config.addr, "env", app.config.env)

	return srv.ListenAndServe()
}