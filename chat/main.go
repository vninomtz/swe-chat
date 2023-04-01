package main

import (
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/stretchr/gomniauth"
	"github.com/stretchr/gomniauth/providers/github"
	"github.com/stretchr/objx"
	"github.com/stretchr/signature"
	"github.com/vninomtz/swe-chat/trace"
)

type templateHandler struct {
	once     sync.Once
	filename string
	tmpl     *template.Template
}

func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.once.Do(func() {
		t.tmpl = template.Must(template.ParseFiles(filepath.Join("./chat/templates", t.filename)))
	})
	data := map[string]interface{}{
		"Host": r.Host,
	}
	if authCookie, err := r.Cookie("auth"); err == nil {
		data["UserData"] = objx.MustFromBase64(authCookie.Value)
	}
	t.tmpl.Execute(w, data)
}

func main() {
	var addr = flag.String("addr", ":8080", "The addr of the application")
	flag.Parse()
	// setup gomniauth
	gomniauth.SetSecurityKey(signature.RandomKey(64))
	// get from .env
	gomniauth.WithProviders(
		github.New(
			"GITHUB_CLIENT_ID",
			"GITHUB_CLIENT_SECRET",
			"http://localhost:8080/auth/callback/github"),
	)
	r := newRoom()
	r.tracer = trace.New(os.Stdout)
	handler := templateHandler{
		filename: "chat.html",
	}
	loginHdl := templateHandler{filename: "login.html"}
	http.Handle("/chat", MustAuth(&handler))
	http.Handle("/login", &loginHdl)
	http.HandleFunc("/auth/", loginHandler)
	http.Handle("/room", r)

	go r.run()

	log.Println("Starting web server on", *addr)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
