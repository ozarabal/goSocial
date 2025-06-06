package main

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/ozarabal/goSocial/internal/store"
)
type CreatePostPayload struct {
	Title 	string 		`json:"title" validate:"required,max=100"`
	Content string 		`json:"content" validate:"required,max=1000"`
	Tags	[]string 	`json:"tags"`
}


type contextKey string
const postKey contextKey = "post"

// CreatePost godoc
//
//	@Summary		Creates a post
//	@Description	Creates a post
//	@Tags			posts
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		CreatePostPayload	true	"Post payload"
//	@Success		201		{object}	store.Post
//	@Failure		400		{object}	error
//	@Failure		401		{object}	error
//	@Failure		500		{object}	error
//	@Security		ApiKeyAuth
//	@Router			/posts [post]
func (app *application) createPostHandler(w http.ResponseWriter, r *http.Request){
	// read from request
	var payload CreatePostPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w,r,err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w,r,err)
		return
	}

	user := getUserFromContext(r)

	post := &store.Post{
		Title: payload.Title,
		Content: payload.Content,
		Tags: payload.Tags,
		UserId: user.ID,
	}
	ctx := r.Context()

	// ask internal to store data to database
	if err := app.store.Posts.Create(ctx, post); err != nil {
		app.internalServerError(w, r, err)
	}

	if err := app.jsonResponse(w, http.StatusCreated, post); err != nil{
		app.internalServerError(w, r, err)
	}
}

// GetPost godoc
//
//	@Summary		Fetches a post
//	@Description	Fetches a post by ID
//	@Tags			posts
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"Post ID"
//	@Success		200	{object}	store.Post
//	@Failure		404	{object}	error
//	@Failure		500	{object}	error
//	@Security		ApiKeyAuth
//	@Router			/posts/{id} [get]
func (app *application) getPostHandler(w http.ResponseWriter, r *http.Request){
	// get id from url paramater
	post := getPostFromCtx(r)

	comments, err := app.store.Comments.GetCommentsByPostID(r.Context(), post.ID)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	post.Comments = comments

	if err = app.jsonResponse(w, http.StatusCreated, post); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

// DeletePost godoc
//
//	@Summary		Deletes a post
//	@Description	Delete a post by ID
//	@Tags			posts
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"Post ID"
//	@Success		204	{object} string
//	@Failure		404	{object}	error
//	@Failure		500	{object}	error
//	@Security		ApiKeyAuth
//	@Router			/posts/{id} [delete]
func (app *application) deletePostHandler(w http.ResponseWriter, r *http.Request){
	idParam := chi.URLParam(r, "postID")
	id, err := strconv.ParseInt(idParam, 10, 64)

	if err != nil{
		app.internalServerError(w, r, err)
		return
	}

	ctx := r.Context()

	if err:= app.store.Posts.DeleteByID(ctx, id); err != nil{
		switch {
		case errors.Is(err, store.ErrorNotFound):
			app.notFoundResponse(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
	}
	w.WriteHeader(http.StatusNoContent)
}

type UpdatePostPayload struct {
	Title 	*string	`json:"title" validate:"omitempty"`
	Content *string 	`json:"content" validate:"omitempty"`
}

// UpdatePost godoc
//
//	@Summary		Updates a post
//	@Description	Updates a post by ID
//	@Tags			posts
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int					true	"Post ID"
//	@Param			payload	body		UpdatePostPayload	true	"Post payload"
//	@Success		200		{object}	store.Post
//	@Failure		400		{object}	error
//	@Failure		401		{object}	error
//	@Failure		404		{object}	error
//	@Failure		500		{object}	error
//	@Security		ApiKeyAuth
//	@Router			/posts/{id} [patch]
func (app *application) updatePostHandler(w http.ResponseWriter, r *http.Request){
	post:= getPostFromCtx(r)

	var payload UpdatePostPayload
	if err := readJSON(w, r, &payload); err != nil{
		app.badRequestResponse(w,r, err)
		return
	} 

	if err := Validate.Struct(payload); err != nil{
		app.badRequestResponse(w,r, err)
		return
	}

	if payload.Content != nil {
		post.Content = *payload.Content
	}
	if payload.Title != nil {
		post.Title = *payload.Title
	}

	ctx := r.Context()
	if err := app.updatePost(ctx, post); err != nil {
		app.internalServerError(w,r,err)
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, post); err != nil {
		app.internalServerError(w, r, err)
	}
	
}

func (app *application) postContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){

	idParam := chi.URLParam(r, "postID")
	id, err := strconv.ParseInt(idParam, 10, 64)

	if err != nil{
		app.internalServerError(w, r, err)
		return
	}

	ctx := r.Context()
	
	post, err := app.store.Posts.GetByID(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrorNotFound):
			app.notFoundResponse(w,r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	ctx = context.WithValue(ctx, postKey, post)
	next.ServeHTTP(w, r.WithContext(ctx))

	})
}

func getPostFromCtx(r *http.Request) *store.Post {
	post, _ := r.Context().Value("post").(*store.Post)
	return post
}

func (app *application) updatePost(ctx context.Context, post *store.Post) error {
	if err := app.store.Posts.Update(ctx, post); err != nil{
		return err
	}

	app.cacheStorage.Users.Delete(ctx, post.UserId)
	return nil
}