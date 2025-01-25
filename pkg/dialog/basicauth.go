package dialog

import (
	"encoding/base64"
	"net/http"
	"strings"
)

func BasicAuth(authUsername, authPassword string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			askForCredentials(w)
			return
		}

		authParts := strings.SplitN(authHeader, " ", 2)
		if len(authParts) != 2 || authParts[0] != "Basic" {
			askForCredentials(w)
			return
		}

		payload, err := base64.StdEncoding.DecodeString(authParts[1])
		if err != nil {
			askForCredentials(w)
			return
		}

		pair := strings.SplitN(string(payload), ":", 2)
		if len(pair) != 2 ||
			pair[0] != authUsername ||
			pair[1] != authPassword {
			askForCredentials(w)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func askForCredentials(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
	http.Error(w, "Unauthorized", http.StatusUnauthorized)
}
