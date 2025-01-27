package main

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/ozarabal/goSocial/internal/store"
)

type application struct {
	config 	config
	store 	store.Storage
}

type config struct {
	addr string
	db dbConfig
	env string
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


	})	

	return r
}

func (app *application) run(mux http.Handler)error {

	srv := &http.Server{
		Addr : app.config.addr,
		Handler: mux,
		ReadTimeout: time.Second * 10,
		WriteTimeout: time.Second * 30,
		IdleTimeout: time.Minute,
	}

	log.Printf("server has started at %s", app.config.addr)

	return srv.ListenAndServe()
}