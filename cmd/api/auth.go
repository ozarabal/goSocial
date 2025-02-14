package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/ozarabal/goSocial/internal/mailer"
	"github.com/ozarabal/goSocial/internal/store"
)

type RegisterUserPayload struct {
	Username string `json:"username" validate:"required,max=100"`
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=3,max=72"`
}

type UserWithToken struct{
	*store.User
	Token string `json:"token"`
}

// registeredUserHandler godoc
//
//	@Summary		Register a user
//	@Description	Register a user
//	@Tags			authentication
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		RegisterUserPayload	true	"User credentials"
//	@Success		201		{object}	UserWithToken		"User registered"
//	@Failure		400		{object}	error
//	@Failure		500		{object}	error
//	@Router			/authentication/user [post]
func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request){
	var payload RegisterUserPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w,r,err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w,r,err)
		return
	}

	user := &store.User{
		Username: payload.Username,
		Email: payload.Email,
	}

	// hash the user payload
	if err := user.Password.Set(payload.Password); err != nil{
		app.internalServerError(w, r, err)
		return
	}
	app.logger.Infow("payload valid", "payload", payload)
	app.logger.Infow("storing user", "user", user)

	ctx := r.Context()

	plainToken := uuid.New().String()

	// store token
	hash := sha256.Sum256([]byte(plainToken))
	hashToken := hex.EncodeToString(hash[:])
	// store user
	err := app.store.Users.CreateAndInvite(ctx, user, hashToken,app.config.mail.exp)
	if err != nil {
		switch err {
		case store.ErrDuplicateEmail:
			app.badRequestResponse(w,r,err)
		case store.ErrDuplicateUsername:
			app.badRequestResponse(w,r,err)
		default:
			app.internalServerError(w,r,err)	
		}
		return
	}

	userWithToken := UserWithToken{
		User: user,
		Token: plainToken,
	}

	activationURL := fmt.Sprintf("%s/confirm/%s", app.config.frontendURL, plainToken)

	isProdEnv := app.config.env == "production"
	vars := struct{
		Username string 
		ActivationURL string
		}{
			Username: user.Username,
			ActivationURL: activationURL,
		}
	
	// send email
	status, err := app.mailer.Send(mailer.UserWelcomeTemplate, user.Username, user.Email, vars, !isProdEnv)
	if err != nil{
		app.logger.Errorw("error sending welcome email", "error", err)
		
		//rollback
		if err := app.store.Users.Delete(ctx, user.ID); err != nil {
			app.logger.Errorw("error deleting user", "error", err)
		}

		app.internalServerError(w,r,err)
		return
	}

	app.logger.Infow("Email sent", "status code", status)

	if err := app.jsonResponse(w, http.StatusCreated, userWithToken); err != nil {
		app.internalServerError(w, r, err)
		return 
	}
}

type CreateUserTokenPayload struct{
	Email		string	`json:"email" validate:"required,email,max=255"`
	Password	string	`json:"password" validate:"required,min=3,max=72"`
}


// createTokenHandler godoc
//
//	@Summary		Creates a token
//	@Description	Creates a token for a user
//	@Tags			authentication
//	@Accept			json
//	@Produce		json
//	@Param			payload					body		CreateUserTokenPayload	true	"User credentials"
//	@Success		200						{string}	string					"Token"
//	@Failure		400						{object}	error
//	@Failure		401						{object}	error
//	@Failure		500						{object}	error
//	@Router			/authentication/token	[post]
func (app *application) createTokenHandler(w http.ResponseWriter, r *http.Request)	{
	//parse payload
	var payload CreateUserTokenPayload
	if err := readJSON(w, r, &payload); err != nil{
		app.badRequestResponse(w, r,err)
		return
	}

	if err := Validate.Struct(payload); err != nil{
		app.badRequestResponse(w,r,err)
		return
	}

	// fetch the user (check if the user exist) from the payload
	user, err := app.store.Users.GetByEmail(r.Context(), payload.Email)
	if err != nil{
		switch err {
		case store.ErrorNotFound:
			app.unauthorizedErrorResponse(w,r,err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	// generate the token -> add claims
	claims := jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(app.config.auth.token.exp).Unix(),
		"iat": time.Now().Unix(),
		"nbf": time.Now().Unix(),
		"iss": app.config.auth.token.iss,
		"aud": app.config.auth.token.iss,
	}

	token, err := app.authenticator.GenerateToken(claims)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	// send it to the client
	if err := app.jsonResponse(w,http.StatusCreated, token); err != nil {
		app.internalServerError(w,r,err)
	}
}