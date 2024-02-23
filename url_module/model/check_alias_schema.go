package model

import (
	"codedln/shared/http_error"
	"codedln/util/constant"
	"net/http"
)

type CheckAliasSchema struct {
	Alias string `json:"alias"`
}

func NewCheckAliasSchema() CheckAliasSchema {
	return CheckAliasSchema{}
}

func (s CheckAliasSchema) Validate() error {
	if len(s.Alias) < constant.AliasMinLength || len(s.Alias) > constant.AliasMaxLength {
		return http_error.New(http.StatusBadRequest, "alias length must be within 3 and 8 characters")
	}
	return nil
}
