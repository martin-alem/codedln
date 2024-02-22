package model

import (
	"codedln/shared/http_error"
	"codedln/util/constant"
	"net/http"
	"regexp"
)

var validUrl = regexp.MustCompile(`((http|https)://)(www.)?[a-zA-Z0-9@:%._\\+~#?&/=]{2,256}\\.[a-z]{2,6}\\b([-a-zA-Z0-9@:%._\\+~#?&/=]*)`)

type CreateUrlSchema struct {
	OriginalUrl string
	Alias       string
}

func NewCreateUrlSchema() CreateUrlSchema {
	return CreateUrlSchema{}
}

func (s CreateUrlSchema) Validate() error {
	if s.Alias != "" && len(s.Alias) > constant.AliasMaxLength {
		return http_error.New(http.StatusBadRequest, "alias length must be within 8 characters")
	}

	if !validUrl.MatchString(s.OriginalUrl) {
		return http_error.New(http.StatusBadRequest, "invalid url")
	}

	return nil
}
