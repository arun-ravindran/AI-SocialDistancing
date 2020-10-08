package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"text/template"

	"github.com/stretchr/gomniauth"
	"github.com/stretchr/gomniauth/providers/github"
	"github.com/stretchr/objx"
	"github.com/stretchr/signature"

)


// templ represents a single template
type templateHandler struct {
	filename string
	templ    *template.Template
}

// ServeHTTP handles the HTTP request.
func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if t.templ == nil {
		t.templ = template.Must(template.ParseFiles(filepath.Join("templates", t.filename)))
	}

	data := map[string]interface{}{
		"Host": r.Host,
	}
	if authCookie, err := r.Cookie("auth"); err == nil {
		data["UserData"] = objx.MustFromBase64(authCookie.Value)
	}

	t.templ.Execute(w, data)
}

var addr = flag.String("host", ":8080", "website address")

func main() {

	// setup gomniauth
    gomniauth.SetSecurityKey(signature.RandomKey(64)) // Random key each time the application starts
    gomniauth.WithProviders(
        github.New(os.Getenv("GITHUB_CLIENT_ID"), os.Getenv("GITHUB_CLIENT_SECRET"), "http://localhost:8080/auth/callback/github"),
    )


	mux := http.NewServeMux()
	mux.Handle("/", MustAuth(&templateHandler{filename: "view.html"})) // Calls login if auth cookie not set; else goes to server
	mux.Handle("/login", &templateHandler{filename: "login.html"}) // Calls auth for authentication
	mux.HandleFunc("/auth/", loginHandler) // Returns to view after authentication
	mux.HandleFunc("/server", sceneHandler) // Streams images and key points through websocket
	//mux.Handle("/view", MustAuth(&templateHandler{filename: "view.html"}))

	mux.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{
			Name:   "auth",
			Value:  "",
			Path:   "/",
			MaxAge: -1,
		})
		w.Header().Set("Location", "/")
		w.WriteHeader(http.StatusTemporaryRedirect)
	})
/*
	http.Handle("/avatars/",
		http.StripPrefix("/avatars/",
			http.FileServer(http.Dir("./avatars"))))

*/
	// start the web server
	log.Println("Starting web server on", *addr)
	if err := http.ListenAndServe(*addr, mux); err != nil {
		log.Fatal("ListenAndServe:", err)
	}

}

