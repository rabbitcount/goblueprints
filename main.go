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
//	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
//		w.Write([]byte(`
//			<html>
//				<head>
//					<title>Chat</title>
//				</head>
//				<body>
//					Let's chat!
//				</body>
//			</html>
//		`))
//	})
	// root use own defined handler
	// We do not store a reference to our newly created templateHandler type,
	// but that's OK because we don't need to refer to it again.
	http.Handle("/", &templateHandler{filename: "chat.html"})
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}



