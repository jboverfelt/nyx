package main

import (
	"log"
	"net/http"

	"golang.org/x/oauth2"
)

// Error represents a handler error. It provides methods for a HTTP status
// code and embeds the built-in error interface.
type Error interface {
	error
	Status() int
}

// StatusError represents an error with an associated HTTP status code.
type StatusError struct {
	Code int
	Err  error
}

// Allows StatusError to satisfy the error interface.
func (se StatusError) Error() string {
	return se.Err.Error()
}

// Status returns our HTTP status code.
func (se StatusError) Status() int {
	return se.Code
}

// Env represents handler dependencies
type Env struct {
	DB        Store
	OAuthConf *oauth2.Config
}

// Handler represents an HTTP handler that can return errors
// and access dependencies in a type-safe way
type Handler struct {
	*Env
	h func(e *Env, w http.ResponseWriter, r *http.Request) error
}

// NewHandler allocates a new handler
func NewHandler(e *Env, handlerFunc func(e *Env, w http.ResponseWriter, r *http.Request) error) Handler {
	return Handler{e, handlerFunc}
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := h.h(h.Env, w, r)
	if err != nil {
		switch e := err.(type) {
		case Error:
			// We can retrieve the status here and write out a specific
			// HTTP status code.
			log.Printf("HTTP %d - %s", e.Status(), e)
			http.Error(w, e.Error(), e.Status())
		default:
			// Any error types we don't specifically look out for default
			// to serving a HTTP 500
			http.Error(w, http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError)
		}
	}
}
