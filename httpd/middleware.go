package httpd

import (
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/adamveld12/muxwrap"
)

func BasicAuth(auth func(username, pass string) error) muxwrap.Middleware {
	return muxwrap.Middleware(func(next http.Handler) http.Handler {

		return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			res.Header().Set("WWW-Authenticate", `Basic realm="Goku"`)

			s := strings.SplitN(req.Header.Get("Authorization"), " ", 2)
			if len(s) != 2 {
				http.Error(res, "Not authorized", 401)
				return
			}

			b, err := base64.StdEncoding.DecodeString(s[1])
			if err != nil {
				http.Error(res, err.Error(), 401)
				return
			}

			pair := strings.SplitN(string(b), ":", 2)
			if len(pair) != 2 {
				http.Error(res, "Not authorized", 401)
				return
			}

			if err := auth(pair[0], pair[1]); err != nil {
				http.Error(res, "Not authorized", 401)
				return
			}

			next.ServeHTTP(res, req)
		})

	})
}
