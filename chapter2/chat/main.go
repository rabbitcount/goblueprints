package main

import (
	"log"
	"net/http"
	"sync"
	"text/template"
	"path/filepath"
	"flag"
	"os"
	"github.com/rabbitcount/goblueprints/chapter1/trace"
	"github.com/stretchr/gomniauth"
	"github.com/stretchr/gomniauth/providers/facebook"
	"github.com/stretchr/gomniauth/providers/github"
	"github.com/stretchr/gomniauth/providers/google"
	"github.com/stretchr/objx"
)

// templ represents a single template
type templateHandler struct {
	// guarantees that the function we pass as an argument will only be executed once
	once     sync.Once
	filename string
	templ    *template.Template
}

// ServeHTTP is a function of Handler.
// ServeHTTP handles the HTTP request.
// load the source  le, compile the template and execute it, and write
// the output to the speci ed http.ResponseWriter object
func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.once.Do(func() {
		t.templ = template.Must(template.ParseFiles(filepath.Join("chapter2/chat/templates", t.filename)))
	})
	data := map[string]interface{} {
		"Host": r.Host,
	}
	if authCookie, err := r.Cookie("auth"); err == nil {
		// auth in cookie is username which encode by base64
		data["UserData"] = objx.MustFromBase64(authCookie.Value)
	}
	t.templ.Execute(w, data)
}

func main()  {
	// The de nition for the addr variable sets up our flag
	// as a string that defaults to :8080
	addr := flag.String("addr", ":8080", "The addr of the application.")
	// must call flag. Parse() that parses the arguments
	// and extracts the appropriate information.
	// Then, we can reference the value of the host  ag by using *addr.
	flag.Parse() // parse the flags

	// set up gomniauth
	gomniauth.SetSecurityKey("some long key")
	gomniauth.WithProviders(
		facebook.New("key", "secret", ""),
		github.New("key", "secret", ""),
		google.New("151570833065-i9p63mogjm7adt0h0490or9bvqua0r2l.apps.googleusercontent.com",
			"jzEEORYarixD30S6qyosGrWe",
			"http://localhost:3000/auth/callback/google"),
	)

	r := newRoom()
	r.tracer = trace.New(os.Stdout)

	http.Handle("/chat", MustAuth(&templateHandler{filename: "chat.html"}))
	http.Handle("/login", &templateHandler{filename: "login.html"})
	http.HandleFunc("/auth/", loginHandler)
	http.Handle("/room", r)
	// get the room going
	// running the room in a separate Go routine
	// so that the chatting operations occur in the background,
	// allowing our main thread to run the web server.
	go r.run()
	// start the web server
	log.Println("Starting web server on", *addr)
	// start the web server
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}



