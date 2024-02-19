package types

import "net/http"

type HTTPHandler func(http.ResponseWriter, *http.Request) error

type Middleware func(HTTPHandler) HTTPHandler
