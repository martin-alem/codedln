package controller

import (
	"codedln/shared/http_error"
	"codedln/url_module/model"
	"codedln/url_module/service"
	"codedln/util/constant"
	"codedln/util/helpers"
	"codedln/util/types"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
	"strings"
)

type UrlController struct {
	urlService *service.UrlService
}

func New(urlService *service.UrlService) *UrlController {
	return &UrlController{
		urlService: urlService,
	}
}

func (c *UrlController) CreateUrl(w http.ResponseWriter, r *http.Request) error {

	UserIDValue := r.Context().Value(constant.AuthUserKey)

	var userId string

	if UserIDValue == nil {
		userId = ""
	} else {
		UserIdPayload, ok := UserIDValue.(types.AuthUser)
		if !ok {
			return http_error.New(http.StatusBadRequest, "invalid user id")
		}
		userId = UserIdPayload.UserId
	}

	UrlValue := r.Context().Value(constant.PayloadKey)

	if UrlValue == nil {
		return http_error.New(http.StatusBadRequest, "unable to get payload")
	}

	UrlPayload, ok := UrlValue.(model.CreateUrlSchema)
	if !ok {
		return http_error.New(http.StatusBadRequest, "invalid url payload")
	}

	url, err := c.urlService.CreateUrl(r.Context(), UrlPayload.OriginalUrl, UrlPayload.Alias, userId)
	if err != nil {
		return err
	}

	return helpers.JSONResponse(w, http.StatusCreated, url)
}

func (c *UrlController) CheckAliasExistence(w http.ResponseWriter, r *http.Request) error {

	AliasValue := r.Context().Value(constant.PayloadKey)

	if AliasValue == nil {
		return http_error.New(http.StatusBadRequest, "no alias provided")
	}

	AliasPayload, ok := AliasValue.(model.CheckAliasSchema)
	if !ok {
		return http_error.New(http.StatusBadRequest, "invalid alias payload")
	}

	err := c.urlService.CheckAliasExistence(r.Context(), AliasPayload.Alias)
	if err != nil {
		return err
	}

	return helpers.JSONResponse(w, http.StatusOK, nil)
}

func (c *UrlController) GetUrls(w http.ResponseWriter, r *http.Request) error {
	UserIDValue := r.Context().Value(constant.AuthUserKey)

	if UserIDValue == nil {
		return http_error.New(http.StatusBadRequest, "no user id")
	}
	UserIdPayload, ok := UserIDValue.(types.AuthUser)
	if !ok {
		return http_error.New(http.StatusBadRequest, "invalid user id")
	}

	query := r.URL.Query()

	var sort types.DateSort
	var limit int64

	search := query.Get("query")
	sortStr := query.Get("date_sort")

	switch sortStr {
	case "-1":
		sort = constant.NewestDate
		break
	case "1":
		sort = constant.OldestDate
		break
	default:
		sort = constant.NewestDate
	}

	limitStr := query.Get("limit")
	val, err := strconv.ParseInt(limitStr, 10, 64)
	if err != nil {
		limit = constant.MaxLimit
	} else {
		limit = val
	}

	result, err := c.urlService.GetUrls(r.Context(), search, sort, limit, UserIdPayload.UserId)
	if err != nil {
		return err
	}

	return helpers.JSONResponse(w, http.StatusOK, result)

}

func (c *UrlController) GetUrl(w http.ResponseWriter, r *http.Request) error {
	UserIDValue := r.Context().Value(constant.AuthUserKey)

	if UserIDValue == nil {
		return http_error.New(http.StatusBadRequest, "no user id")
	}
	UserIdPayload, ok := UserIDValue.(types.AuthUser)
	if !ok {
		return http_error.New(http.StatusBadRequest, "invalid user id")
	}

	vars := mux.Vars(r)
	fmt.Println("RAW: ", vars)
	urlId, exist := vars["urlId"]
	if !exist {
		return http_error.New(http.StatusBadRequest, "no url id found")
	}

	url, err := c.urlService.GetUrl(r.Context(), urlId, UserIdPayload.UserId)
	if err != nil {
		return err
	}

	return helpers.JSONResponse(w, http.StatusOK, url)
}

func (c *UrlController) DeleteUrl(w http.ResponseWriter, r *http.Request) error {
	UserIDValue := r.Context().Value(constant.AuthUserKey)

	if UserIDValue == nil {
		return http_error.New(http.StatusBadRequest, "no user id")
	}
	UserIdPayload, ok := UserIDValue.(types.AuthUser)
	if !ok {
		return http_error.New(http.StatusBadRequest, "invalid user id")
	}

	vars := mux.Vars(r)
	urlId, exist := vars["urlId"]
	if !exist {
		return http_error.New(http.StatusBadRequest, "no url id found")
	}

	err := c.urlService.DeleteUrl(r.Context(), urlId, UserIdPayload.UserId)
	if err != nil {
		return err
	}

	return helpers.JSONResponse(w, http.StatusOK, nil)
}

func (c *UrlController) DeleteUrls(w http.ResponseWriter, r *http.Request) error {
	UserIDValue := r.Context().Value(constant.AuthUserKey)

	if UserIDValue == nil {
		return http_error.New(http.StatusBadRequest, "no user id")
	}
	UserIdPayload, ok := UserIDValue.(types.AuthUser)
	if !ok {
		return http_error.New(http.StatusBadRequest, "invalid user id")
	}

	query := r.URL.Query()
	url := query.Get("url")
	if url == "" {
		return http_error.New(http.StatusBadRequest, "one url id must be provided")
	}

	urls := strings.Split(url, ",")

	err := c.urlService.DeleteUrls(r.Context(), urls, UserIdPayload.UserId)
	if err != nil {
		return err
	}

	return helpers.JSONResponse(w, http.StatusOK, nil)
}

func (c *UrlController) Redirect(w http.ResponseWriter, r *http.Request) error {
	query := r.URL.Query()
	alias := query.Get("alias")
	if alias == "" {
		return http_error.New(http.StatusBadRequest, "no alias found")
	}

	originalUrl, err := c.urlService.Redirect(r.Context(), alias)
	if err != nil {
		return err
	}

	http.Redirect(w, r, originalUrl, http.StatusTemporaryRedirect)
	return nil
}
