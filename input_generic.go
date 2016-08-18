package main

import (
	"net/http"

	"github.com/gorilla/mux"
)

func init() {
	registerInputHandler("generic", InputGeneric{})
}

type InputGeneric struct{}

func (i InputGeneric) RegisterOnRouter(r *mux.Router) {
	r.HandleFunc("/{type}", i.handle)
}

func (i InputGeneric) GetHelp() inputHandlerHelp {
	return inputHandlerHelp{
		Path:        "/generic/<type>",
		Description: "POST with key-value pairs as application/x-www-form-urlencoded",
	}
}

func (i InputGeneric) handle(res http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	m := createMessage(vars["type"])

	r.ParseForm()

	for k, v := range r.Form {
		if len(v) > 1 {
			http.Error(res, "Mutli-value keys are not supported.", http.StatusBadRequest)
			return
		}
		m.Set(k, v[0])
	}

	if err := enqueueMessage(m); err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	res.WriteHeader(http.StatusCreated)
}
