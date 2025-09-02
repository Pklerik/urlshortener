package middleware

import "net/http"

// ApplyMiddleware provide Middleware construct for all middlewares.
func ApplyMiddleware(targetFunc http.HandlerFunc, mdFuncs ...func(h http.HandlerFunc) http.HandlerFunc) http.HandlerFunc {
	for i := len(mdFuncs) - 1; i >= 0; i-- {
		targetFunc = mdFuncs[i](targetFunc)
	}

	return targetFunc
}
