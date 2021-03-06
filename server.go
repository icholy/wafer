package main

import (
	"encoding/json"
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"
)

var (
	signingKey []byte
	config     Config

	configFile = flag.String("config", "config.json", "Configuration file")
)

func init() {

	flag.Parse()

	// read config file
	f, err := os.Open(*configFile)
	if err != nil {
		log.Fatal(err)
	}
	if err := json.NewDecoder(f).Decode(&config); err != nil {
		log.Fatal(err)
	}

	signingKey = []byte(config.Key)
}

// Authenticate a user
func Authenticate(username string, password string) (*User, error) {
	user := &User{0, "Ilia Choly", []string{"dev", "admin"}}
	return user, nil
}

func main() {

	tmpl := template.Must(template.ParseGlob("templates/*.tmpl"))

	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {

		// only accept GET requests
		if r.Method != "GET" {
			http.Error(w, http.StatusText(405), 405)
			return
		}

		var (
			redirect = r.URL.Query().Get("redirect")
			appname  = r.URL.Query().Get("appname")
		)

		if redirect == "" {
			redirect = r.Referer()
		}

		// render template
		data := struct {
			Redirect string
			AppName  string
		}{redirect, appname}

		if err := tmpl.ExecuteTemplate(w, "login.html.tmpl", &data); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

	})

	http.HandleFunc("/api/login", func(w http.ResponseWriter, r *http.Request) {

		// only accept POST requests
		if r.Method != "POST" {
			http.Error(w, http.StatusText(405), 405)
			return
		}

		// authenticate
		var (
			username = r.PostFormValue("username")
			password = r.PostFormValue("password")
		)
		user, err := Authenticate(username, password)
		if err != nil {
			http.Error(w, err.Error(), 401)
			return
		}

		// generate JSON web token
		token, err := user.JWT()
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}

		var (
			redirect = r.PostFormValue("redirect")
			appname  = r.PostFormValue("appname")

			hooks []string
		)

		if app, ok := config.Applications.Lookup(appname); ok {
			// If an application is specified, only invoke its hook
			if redirect == "" || app.ForceRedirect {
				redirect = app.RedirectURL
			}
			hooks = []string{app.WebHookURL}
		} else {
			// if no application is specified, invoke all the hooks
			hooks = config.Applications.WebHooks()
		}

		data := struct {
			JWT      string
			WebHooks []string
			Redirect string
		}{token, hooks, redirect}

		if err := tmpl.ExecuteTemplate(w, "postlogin.html.tmpl", &data); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

	})

	http.HandleFunc("/api/wafer_hook", func(w http.ResponseWriter, r *http.Request) {

		if r.Method != "GET" {
			http.Error(w, http.StatusText(405), 405)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:  "jwt",
			Value: r.URL.Query().Get("jwt"),
			Path:  "/",
		})

		w.WriteHeader(200)
	})

	if err := http.ListenAndServe(config.ListenAddr, nil); err != nil {
		log.Fatal(err)
	}
}
