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

	accessToken := helpers.CreateJWT(types.AuthUser{UserId: user.ID.String()}, constant.AccessTokenTTL)

	if accessToken == "" {
		return http_error.New(http.StatusInternalServerError, "unable to create jwt")
	}

	cookie := helpers.CreateCook(constant.JwtCookieName, accessToken, constant.AccessTokenTTL)
	http.SetCookie(w, &cookie)

	return helpers.JSONResponse(w, http.StatusCreated, user)
}
