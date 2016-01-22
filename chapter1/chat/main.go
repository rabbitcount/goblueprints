package main

import (
	"log"
	"net/http"
	"sync"
	"path/filepath"
	"flag"
	"os"
	"github.com/rabbitcount/goblueprints/chapter1/trace"
	"text/template"
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
//		cwd, _ := os.Getwd()
//		fmt.Println( filepath.Join( cwd, "templates", t.filename ) )

		t.templ = template.Must(template.ParseFiles(filepath.Join("chapter1/chat/templates", t.filename)))
	})
	t.templ.Execute(w, r)
}

func main()  {
	// The de nition for the addr variable sets up our flag
	// as a string that defaults to :8080
	addr := flag.String("addr", ":8080", "The addr of the application.")
	// must call flag. Parse() that parses the arguments
	// and extracts the appropriate information.
	// Then, we can reference the value of the host  ag by using *addr.
	flag.Parse() // parse the flags
	r := newRoom()
	r.tracer = trace.New(os.Stdout)

	http.Handle("/", &templateHandler{filename: "chat.html"})
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



