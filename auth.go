package main

import (
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"golang.org/x/oauth2"

	"github.com/satori/go.uuid"
)

// User represents a user of the application
type User struct {
	// id
	State string
	Email string
	Token string
}

func index(e *Env, w http.ResponseWriter, r *http.Request) error {
	tmpl := template.Must(template.ParseFiles("templates/index.tmpl"))
	state := uuid.NewV4().String()
	return tmpl.Execute(w, struct{ State string }{state})
}

func fitbitLogin(e *Env, w http.ResponseWriter, r *http.Request) error {
	// pull state string out of request along with email and save to db
	state := r.FormValue("state")
	email := r.FormValue("email")

	user := User{
		State: state,
		Email: email,
	}

	err := e.DB.Upsert(user)

	if err != nil {
		return err
	}

	url := e.OAuthConf.AuthCodeURL(state, oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)

	return nil
}

func fitbitCallback(e *Env, w http.ResponseWriter, r *http.Request) error {
	state := r.FormValue("state")
	// lookup state in db and make sure it exists
	u, err := e.DB.GetByState(state)

	if err != nil {
		return err
	} else if u == nil {
		return StatusError{
			Code: http.StatusUnauthorized,
			Err:  errors.New("Invalid CSRF Token"),
		}
	}

	code := r.FormValue("code")
	token, err := e.OAuthConf.Exchange(oauth2.NoContext, code)
	if err != nil {
		return err
	}

	// here, marshal the token to json
	// store json str in db
	tokStr, err := dumpToken(token)

	if err != nil {
		return err
	}

	u.Token = tokStr

	err = e.DB.Upsert(*u)

	if err != nil {
		return err
	}

	log.Printf("Token added: %v", u.Token)

	fmt.Fprintf(w, "Success! Expect an email in the morning!\n")

	return nil
}
