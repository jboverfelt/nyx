package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/fitbit"
)

type config struct {
	ClientID     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
}

var oAuthConf *oauth2.Config

var oAuthState = "state"

func index(w http.ResponseWriter, r *http.Request) {

}

func fitbitLogin(w http.ResponseWriter, r *http.Request) {
	url := oAuthConf.AuthCodeURL(oAuthState, oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func fitbitCallback(w http.ResponseWriter, r *http.Request) {
	state := r.FormValue("state")
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

	oAuthConf = &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		Scopes:       []string{"profile", "sleep"},
		Endpoint:     fitbit.Endpoint,
	}

	http.HandleFunc("/", index)
	http.HandleFunc("/login", fitbitLogin)
	http.HandleFunc("/fitbitCallback", fitbitCallback)
	fmt.Print("Started running on http://127.0.0.1:8080\n")
	fmt.Println(http.ListenAndServe(":8080", nil))
}
