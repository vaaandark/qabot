package dialog

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"os"
	"strings"
)

type User struct {
	Name     string   `json:"name"`
	Password string   `json:"password"`
	Allowed  []string `json:"allowed,omitempty"`
	Welcome  string   `json:"welcome,omitempty"`
}

type Auth struct {
	Admins    []User `json:"admins"`
	NonAdmins []User `json:"non_admins"`
}

func LoadAuthFromFile(path string) (*Auth, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	auth := &Auth{}
	err = json.Unmarshal(bytes, auth)
	if err != nil {
		return nil, err
	}

	return auth, nil
}

func (a Auth) isAdmin(name, password string) bool {
	for _, admin := range a.Admins {
		if name == admin.Name && password == admin.Password {
			return true
		}
	}
	return false
}

func (a Auth) getUser(name string) *User {
	for _, a := range a.Admins {
		if name == a.Name {
			return &a
		}
	}
	for _, na := range a.NonAdmins {
		if name == na.Name {
			return &na
		}
	}
	return nil
}

func (a Auth) isNonAdmin(name, password string) bool {
	for _, nonAdmin := range a.NonAdmins {
		if name == nonAdmin.Name && password == nonAdmin.Password {
			return true
		}
	}
	return false
}

func (a Auth) Auth(name, password string) bool {
	return a.isAdmin(name, password) || a.isNonAdmin(name, password)
}

func getNameAndPassword(r *http.Request) (*string, *string) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, nil
	}

	authParts := strings.SplitN(authHeader, " ", 2)
	if len(authParts) != 2 || authParts[0] != "Basic" {
		return nil, nil
	}

	payload, err := base64.StdEncoding.DecodeString(authParts[1])
	if err != nil {
		return nil, nil
	}

	pair := strings.SplitN(string(payload), ":", 2)
	if len(pair) != 2 {
		return nil, nil
	}

	return &pair[0], &pair[1]
}

func BasicAuth(auth *Auth, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name, password := getNameAndPassword(r)
		if name == nil || password == nil {
			askForCredentials(w)
			return
		}

		if !auth.Auth(*name, *password) {
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
