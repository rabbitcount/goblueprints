package main

import (
	"log"
	"net/http"
	"sync"
"text/template"
	"path/filepath"
)

// templ represents a single template
type templateHandler struct {
	// guarantees that the function we pass as an argument will only be executed once
	once     sync.Once
	filename string
	templ    *template.Template
}

// ServeHTTP handles the HTTP request.
// load the source  le, compile the template and execute it, and write
// the output to the speci ed http.ResponseWriter object
func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.once.Do(func() {
		t.templ = template.Must(template.ParseFiles(filepath.Join("templates", t.filename)))
	})
	t.templ.Execute(w, nil)
}

func main()  {
	r := newRoom()
	http.Handle("/", &templateHandler{filename: "chat.html"})
	http.Handle("/room", r)
	// get the room going
	// running the room in a separate Go routine
	// so that the chatting operations occur in the background,
	// allowing our main thread to run the web server.
	go r.run()
	// start the web server
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}



