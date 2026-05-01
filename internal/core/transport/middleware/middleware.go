package core_http_middleware

import "net/http"

type Middleware func(http.Handler) http.Handler

func ChainMiddleware(next http.Handler, middlewares ...Middleware) http.Handler {
	if len(middlewares) == 0 {
		return next
	}
	for i := len(middlewares) - 1; i >= 0; i-- {
		next = middlewares[i](next)
	}
	return next
}
