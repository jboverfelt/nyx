package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/jboverfelt/nyx/models"
	"github.com/jboverfelt/nyx/store"
	"github.com/satori/go.uuid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/fitbit"
)

type config struct {
	ClientID     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
}

func index(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("index.tmpl"))
	state := uuid.NewV4().String()
	tmpl.Execute(w, struct{ State string }{state})
}

func fitbitLogin(oAuthConf *oauth2.Config, s store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// pull state string out of request along with email and save to db
		state := r.FormValue("state")
		email := r.FormValue("email")

		user := models.User{
			State: state,
			Email: email,
		}

		err := s.Upsert(user)

		if err != nil {
			http.Error(w, "Error saving user", http.StatusInternalServerError)
			return
		}

		url := oAuthConf.AuthCodeURL(state, oauth2.AccessTypeOffline)
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	}
}

func fitbitCallback(w http.ResponseWriter, r *http.Request) {
	state := r.FormValue("state")
	// lookup state in db and make sure it exists
	if state != oAuthState {
		fmt.Printf("invalid oauth state, expected '%s', got '%s'\n", oAuthState, state)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	code := r.FormValue("code")
	token, err := oAuthConf.Exchange(oauth2.NoContext, code)
	if err != nil {
		fmt.Printf("oauthConf.Exchange() failed with '%s'\n", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// here, marshal the token to json
	// store json str in db
	//

	client := oAuthConf.Client(oauth2.NoContext, token)

	fmt.Println(client)

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func hydrateConfig() config {
	cfgFile, err := os.Open("config.json")
	defer cfgFile.Close()

	if err != nil {
		log.Fatalf("Could not open config.json file: %v", err)
	}

	var cfg config
	err = json.NewDecoder(cfgFile).Decode(&cfg)

	if err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	return cfg
}

func main() {
	cfg := hydrateConfig()

	oAuthConf := &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		Scopes:       []string{"profile", "sleep"},
		Endpoint:     fitbit.Endpoint,
	}

	store := store.NewInMemoryStore()

	http.HandleFunc("/", index)
	http.HandleFunc("/login", fitbitLogin(oAuthConf, store))
	http.HandleFunc("/fitbitCallback", fitbitCallback)
	fmt.Print("Started running on http://127.0.0.1:8080\n")
	fmt.Println(http.ListenAndServe(":8080", nil))
}
