package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/dgrijalva/jwt-go"
)

const signingKey = "secret"

// WebHooks is a list
var WebHooks = []string{
	"http://sub1.domain.com/wafer_hook",
	"http://sub2.domain.com/wafer_hook",
}

// User info
type User struct {
	ID     int64    `json:"id"`
	Name   string   `json:"name"`
	Groups []string `json:"groups"`
}

// JWT converts the user to a JSON Web Token
func (u *User) JWT() (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	token.Claims = map[string]interface{}{
		"id":     u.ID,
		"name":   u.Name,
		"groups": u.Groups,
	}
	return token.SignedString(signingKey)
}

// Authenticate a user
func Authenticate(username string, password string) (*User, error) {
	user := &User{0, "Ilia Choly", []string{"dev", "admin"}}
	return user, nil
}

func main() {

	http.HandleFunc("/api/login", func(w http.ResponseWriter, r *http.Request) {
		// authenticate
		var (
			username = r.PostFormValue("username")
			password = r.PostFormValue("password")
		)
		user, err := Authenticate(username, password)
		if err != nil {
			http.Error(w, http.StatusText(401), 401)
			return
		}

		token, err := user.JWT()
		if err != nil {
			http.Error(w, http.StatusText(400), 400)
			return
		}

		fmt.Fprint(w, token)
	})

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}