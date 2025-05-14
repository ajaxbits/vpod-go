package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/urfave/cli/v2"
)

type authData struct {
	User string
	Pass string
}

func NewAuthData(cCtx *cli.Context) (*authData, error) {
	user := cCtx.String("user")
	var pass string
	if cCtx.String("password-file") != "" {
		path := cCtx.String("password-file")
		contents, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		pass = strings.TrimSpace(string(contents))
	} else {
		pass = cCtx.String("password")
	}
	return &authData{
		User: user,
		Pass: pass,
	}, nil
}

func NewBasicAuth(wanted *authData) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return basicAuth(wanted, next)
	}
}

func basicAuth(wanted *authData, next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || !validateCredentials(user, pass, wanted) {
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	}
}

func validateCredentials(username string, password string, wanted *authData) bool {
	return username == wanted.User && password == wanted.Pass
}
