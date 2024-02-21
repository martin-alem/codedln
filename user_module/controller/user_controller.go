package controller

import (
	"codedln/shared/http_error"
	"codedln/user_module/model"
	"codedln/user_module/service"
	"codedln/util/constant"
	"codedln/util/helpers"
	"codedln/util/types"
	"net/http"
)

type UserController struct {
	userService *service.UserService
}

func New(userService *service.UserService) *UserController {
	return &UserController{
		userService: userService,
	}
}

func (c *UserController) CreateUser(w http.ResponseWriter, r *http.Request) error {
	value := r.Context().Value(constant.PayloadKey)

	if value == nil {
		return http_error.New(http.StatusBadRequest, "could not get payload")
	}

	payload, ok := value.(model.CreateUserSchema)
	if !ok {
		return http_error.New(http.StatusBadRequest, "invalid payload")
	}

	user, err := c.userService.CreateUser(r.Context(), payload.IdToken, payload.SignInWith)
	if err != nil {
		return err
	}

	accessToken := helpers.CreateJWT(types.AuthUser{UserId: user.ID.Hex()}, constant.AccessTokenTTL)

	if accessToken == "" {
		return http_error.New(http.StatusInternalServerError, "unable to create jwt")
	}

	cookie := helpers.CreateCookie(constant.JwtCookieName, accessToken, constant.AccessTokenTTL)
	http.SetCookie(w, &cookie)

	return helpers.JSONResponse(w, http.StatusCreated, user)
}

func (c *UserController) GetUser(w http.ResponseWriter, r *http.Request) error {
	value := r.Context().Value(constant.AuthUserKey)

	if value == nil {
		return http_error.New(http.StatusBadRequest, "could not get user id")
	}
	payload, ok := value.(types.AuthUser)
	if !ok {
		return http_error.New(http.StatusBadRequest, "invalid user id")
	}

	user, err := c.userService.GetUser(r.Context(), payload.UserId)
	if err != nil {
		return err
	}

	return helpers.JSONResponse(w, http.StatusOK, user)
}

func (c *UserController) DeleteUser(w http.ResponseWriter, r *http.Request) error {
	value := r.Context().Value(constant.AuthUserKey)

	if value == nil {
		return http_error.New(http.StatusBadRequest, "could not get user id")
	}
	payload, ok := value.(types.AuthUser)
	if !ok {
		return http_error.New(http.StatusBadRequest, "invalid user id")
	}
	err := c.userService.DeleteUser(r.Context(), payload.UserId)

	if err != nil {
		return err
	}

	return helpers.JSONResponse(w, http.StatusOK, nil)
}

func (c *UserController) Logout(w http.ResponseWriter, r *http.Request) error {
	cookie := helpers.CreateCookie(constant.JwtCookieName, "", 0)
	http.SetCookie(w, &cookie)
	return helpers.JSONResponse(w, http.StatusOK, nil)
}
