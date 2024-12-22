package main

import (
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
	"letsgo/ui"
	"net/http"
)

func (app *application) routes() http.Handler {
	r := mux.NewRouter()
	r.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.notFound(w)
	})

	fileServer := http.FileServer(http.FS(ui.Files))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static", fileServer))

	r.HandleFunc("/ping", ping).Methods("GET")

	dynamic := alice.New(app.sessionManager.LoadAndSave, noSurf, app.authenticate)
	r.Handle("/", dynamic.ThenFunc(app.home))
	r.Handle("/snippet/view/{id}", dynamic.ThenFunc(app.snippetView))
	r.Handle("/user/signup", dynamic.ThenFunc(app.userSignup)).Methods("GET")
	r.Handle("/user/signup", dynamic.ThenFunc(app.userSignupPost)).Methods("POST")
	r.Handle("/user/login", dynamic.ThenFunc(app.userLogin)).Methods("GET")
	r.Handle("/user/login", dynamic.ThenFunc(app.userLoginPost)).Methods("POST")

	protected := dynamic.Append(app.requireAuthentication)
	r.Handle("/snippet/create", protected.ThenFunc(app.snippetCreateForm)).Methods("GET")
	r.Handle("/snippet/create", protected.ThenFunc(app.snippetCreatePost)).Methods("POST")
	r.Handle("/user/logout", protected.ThenFunc(app.userLogoutPost)).Methods("POST")

	// middleware chaining
	standard := alice.New(app.recoverPanic, app.logRequest, secureHeaders)
	return standard.Then(r)
}
