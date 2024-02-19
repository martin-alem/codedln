package middleware

import "codedln/util/types"

// ChainMiddlewares applies all provided middlewares to a given handler
func ChainMiddlewares(handler types.HTTPHandler, middlewares ...types.Middleware) types.HTTPHandler {
	for _, middleware := range middlewares {
		handler = middleware(handler)
	}
	return handler
}
