package internalmiddleware

import "net/http"

// ApplyMiddleware provide Middleware construct for all middlewares.
func ApplyMiddleware(targetFunc http.Handler, mdFuncs ...func(next http.Handler) http.Handler) http.Handler {
	for i := len(mdFuncs) - 1; i >= 0; i-- {
		targetFunc = mdFuncs[i](targetFunc)
	}

	return targetFunc
}
