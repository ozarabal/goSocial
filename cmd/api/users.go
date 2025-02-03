package main

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/ozarabal/goSocial/internal/store"
)

type userKey string
const userCtx userKey = "user"

// GetUser godoc
//
//	@Summary		Fetches a user profile
//	@Description	Fetches a user profile by ID
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"User ID"
//	@Success		200	{object}	store.User
//	@Failure		400	{object}	error
//	@Failure		404	{object}	error
//	@Failure		500	{object}	error
//	@Security		ApiKeyAuth
//	@Router			/users/{id} [get]
func (app *application) getUserHandler(w http.ResponseWriter, r *http.Request){
	
	userID, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
	if err != nil {
		app.badRequestResponse(w,r,err)
		return
	}

	user, err := app.getUser(r.Context(), userID)
	if err != nil {
		switch err {
		case store.ErrorNotFound:
			app.notFoundResponse(w,r,err)
			return
		default:
			app.internalServerError(w,r,err)
			return
		}
	}

	if err := app.jsonResponse(w, http.StatusOK, user); err != nil{
		app.internalServerError(w, r, err)
	}
}

// FollowUser godoc
//
//	@Summary		Follows a user
//	@Description	Follows a user by ID
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			userID	path		int		true	"User ID"
//	@Success		204		{string}	string	"User followed"
//	@Failure		400		{object}	error	"User payload missing"
//	@Failure		404		{object}	error	"User not found"
//	@Security		ApiKeyAuth
//	@Router			/users/{userID}/follow [put]
func (app *application) followUserHandler(w http.ResponseWriter, r *http.Request){
	followUser := getUserFromContext(r)
	followedID, err := strconv.ParseInt(chi.URLParam(r,"userID"), 10, 64)
	if err != nil{
		app.badRequestResponse(w,r,err)
		return
	}

	ctx := r.Context()

	if err := app.store.Followers.Follow(ctx, followedID, followUser.ID); err != nil {
		switch err {
		case store.ErrorConflic:
			app.conflictResponse(w,r,err)
			return
		default:
			app.internalServerError(w,r,err)
			return
		}
	}

	if err := app.jsonResponse(w, http.StatusNoContent, nil); err != nil {
		app.internalServerError(w,r,err)
	}
}

// ActivateUser godoc
//
//	@Summary		Activate a user
//	@Description	Activate a user
//	@Tags			users
//	@Produce		json
//	@Param			token	path		string	true	"Invitation token"
//	@Success		204		{string}	string	"User Actived"
//	@Failure		404		{object}	error	"User not found"
//	@Security		ApiKeyAuth
//	@Router			/users/activate/{token} [put]
func (app *application) activateUserHandler(w http.ResponseWriter, r *http.Request){
	token := chi.URLParam(r, "token")

	err:= app.store.Users.Activate(r.Context(), token)
	if err != nil {
		switch err {
		case store.ErrorNotFound:
			app.notFoundResponse(w,r, err)
		default:
			app.internalServerError(w,r, err)
		}
		return
	}

	if err := app.jsonResponse(w,http.StatusNoContent, ""); err != nil{
		app.internalServerError(w, r, err)
	}
}

func (app *application) unfollowUserHandler(w http.ResponseWriter, r *http.Request){
	unfolloweduser := getUserFromContext(r)
	unfolloweduserID, err := strconv.ParseInt(chi.URLParam(r,"userID"), 10, 64)
	if err != nil {
		app.badRequestResponse(w,r,err)
		return 
	}
	
	ctx := r.Context()
	
	if err := app.store.Followers.Unfollow(ctx, unfolloweduserID, unfolloweduser.ID); err != nil{
		app.internalServerError(w,r,err)
		return	
	}

	if err := app.jsonResponse(w, http.StatusNoContent, nil); err != nil {
		app.internalServerError(w,r,err)
	}
}

func getUserFromContext(r *http.Request) *store.User {
	user,_ := r.Context().Value(userCtx).(*store.User)
	return user
}
