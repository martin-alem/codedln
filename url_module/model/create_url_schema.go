package model

import (
	"codedln/shared/http_error"
	"codedln/util/constant"
	"net/http"
	"regexp"
)

var validUrl = regexp.MustCompile(`^(https?://)?([\da-z.-]+)\.([a-z.]{2,6})([/\w .-]*)*/?(\?\S*)?$`)

type CreateUrlSchema struct {
	OriginalUrl string `json:"originalUrl"`
	Alias       string `json:"alias"`
}

func NewCreateUrlSchema() CreateUrlSchema {
	return CreateUrlSchema{}
}

func (s CreateUrlSchema) Validate() error {
	if s.Alias != "" && (len(s.Alias) < constant.AliasMinLength || len(s.Alias) > constant.AliasMaxLength) {
		return http_error.New(http.StatusBadRequest, "alias length must be within 8 characters")
	}

	if !validUrl.MatchString(s.OriginalUrl) {
		return http_error.New(http.StatusBadRequest, "invalid url")
	}

	return nil
}
