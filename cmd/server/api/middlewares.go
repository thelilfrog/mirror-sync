package api

import "net/http"

func recoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			err := recover()
			if err != nil {
				internalServerError(err, w, r)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
