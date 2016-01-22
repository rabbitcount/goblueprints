package main

import (
	"net/http"
	"strings"
	"log"
	"fmt"
	"github.com/stretchr/gomniauth"
	"github.com/stretchr/objx"
)

/**
  * The authHandler type not only implements the ServeHTTP method
  * (which satisfies the http.Handler interface) but also stores (wraps)
  * http.Handler in the next field
 */
type authHandler struct  {
	next http.Handler
}

func (h *authHandler) ServeHTTP(w http.ResponseWriter, r *http.Request){
	if _, err := r.Cookie("auth"); err == http.ErrNoCookie {
		// not authenticated
		w.Header().Set("Location", "/login")
		w.WriteHeader(http.StatusTemporaryRedirect)
	} else if err != nil {
		// some other error
		panic(err.Error())
	} else {
		// success - call the next handler
		h.next.ServeHTTP(w, r)
	}
}

func MustAuth(handler http.Handler) http.Handler {
	return &authHandler{next: handler}
}

// loginHandler handles the third-party login process.
// format: /auth/{action}/{provider}
func loginHandler(w http.ResponseWriter, r *http.Request) {
	segs := strings.Split(r.URL.Path, "/")
	action := segs[2]
	provider := segs[3]
	switch action {
	case "login":
		// get the provider object that matches the object specfied in the URL (such as google or github)
		provider, err := gomniauth.Provider(provider)
		if err != nil {
			log.Fatal("Error when trying to get provider", provider, "-", err)
		}
		// get the location where we must send users in order to start the authentication process
		// GetBeginAuthURL
		// first argument:
		//   is a state map of data that is encoded, and signed and sent to the authentication provider.
		//   The provider doesn't do anything with the state, it just sends it back to our callback endpoint.
		//   This is useful if, for example, we want to redirect the user back to the original page they were
		//   trying to access before the authentication process intervened. For our purpose, we have only the
		//   /chat endpoint, so we don't need to worry about sending any state.
		// second argument:
		//   is a map of additional options that will be sent to the authentication provider,
		//   which somehow modi es the behavior of the authentication process.
		//   For example, you can specify your own scope parameter, which allows you to make a request
		//   for permission to access additional information from the provider. For more information
		//   about the available options, search for OAuth2 on the Internet or read the documentation
		//   for each provider, as these values differ from service to service.
		loginUrl, err := provider.GetBeginAuthURL(nil, nil)
		if err != nil {
			log.Fatalln("Error when trying to GetBeginAuthURL for", provider, "-", err);
		}
		w.Header().Set("Location", loginUrl)
		w.WriteHeader(http.StatusTemporaryRedirect)
	case "callback":
		// get the provider
		provider, err := gomniauth.Provider(provider)
		if err != nil {
			log.Fatalln("Error when trying to get provider", provider, "-", err)
		}

		// exchanging it for an access token as per the OAuth specification
		creds, err := provider.CompleteAuth(objx.MustFromURLQuery(r.URL.RawQuery))
		if err != nil {
			log.Fatalln("Error when trying to complete auth for", provider, "-", err)
		}

		user, err := provider.GetUser(creds)
		if err != nil {
			log.Fatalln("Error when trying to get user from", provider, "-", err)
		}

		authCookieValue := objx.New(map[string]interface{} {
			"name": user.Name(),
		}).MustBase64()
		http.SetCookie(w, &http.Cookie{
			Name: 	"auth",
			Value:	authCookieValue,
			Path:	"/",
		})

		w.Header()["Location"] = []string{"/chat"}
		w.WriteHeader(http.StatusTemporaryRedirect)
	default:
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "Auth action %s not supported", action)
	}
}
