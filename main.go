package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jboverfelt/nyx/models"
	"github.com/jboverfelt/nyx/store"
	"github.com/robfig/cron"
	"github.com/satori/go.uuid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/fitbit"
)

type config struct {
	ClientID     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
	RedirectURL  string `json:"redirectUrl"`
	CronSchedule string `json:"cronSchedule"`
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

func fitbitCallback(oAuthConf *oauth2.Config, s store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		state := r.FormValue("state")
		// lookup state in db and make sure it exists
		u, err := s.GetByState(state)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		} else if u == nil {
			http.Error(w, "Invalid CSRF token", http.StatusUnauthorized)
			return
		}

		code := r.FormValue("code")
		token, err := oAuthConf.Exchange(oauth2.NoContext, code)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// here, marshal the token to json
		// store json str in db
		tokStr, err := json.Marshal(token)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		u.Token = string(tokStr)

		err = s.Upsert(*u)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		log.Printf("Token added: %v", u.Token)

		fmt.Fprintf(w, "Success! Expect an email at 10am!\n")
	}
}

func hydrateConfig() config {
	cfgFile, err := os.Open("config.json")
	defer cfgFile.Close()

	if err != nil {
		log.Fatalf("Could not open config.json file: %v\n", err)
	}

	var cfg config
	err = json.NewDecoder(cfgFile).Decode(&cfg)

	if err != nil {
		log.Fatalf("Error reading config file: %v\n", err)
	}

	return cfg
}

func setupCron(c *cron.Cron, schedule string, oAuthConf *oauth2.Config, s store.Store) {
	urlFirst := "https://api.fitbit.com/1/user/-/sleep/date/"
	urlSecond := ".json?isMainSleep=true"

	log.Printf("Scheduling email func on schedule: %s\n", schedule)

	err := c.AddFunc(schedule, func() {
		log.Println("Starting up cron")
		us, err := s.GetAll()

		if err != nil {
			log.Printf("ERROR: Fetching all users failed: %v\n", err.Error())
		}

		if us == nil {
			return
		}

		for _, u := range us {
			var token oauth2.Token
			err := json.Unmarshal([]byte(u.Token), &token)

			if err != nil {
				log.Printf("ERROR: failed to unmarshal token for email %v\n", u.Email)
				continue
			}

			client := oAuthConf.Client(oauth2.NoContext, &token)

			curDate := time.Now().Format("2006-01-02")

			url := urlFirst + curDate + urlSecond

			resp, err := client.Get(url)

			if err != nil {
				log.Printf("ERROR: fitbit call failed: %v\n", err.Error())
				continue
			}

			defer resp.Body.Close()

			var sleepResponse models.SleepResponse

			err = json.NewDecoder(resp.Body).Decode(&sleepResponse)

			if err != nil {
				log.Printf("ERROR: failed to decode API response: %v", err.Error())
			}

			log.Printf("User %s woke %d times last night \n", u.Email, sleepResponse.Sleep[0].AwakeCount)
		}
	})

	if err != nil {
		log.Fatalf("Could not schedule cron func %v\n", err)
	}
}

func main() {
	cfg := hydrateConfig()

	oAuthConf := &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		Scopes:       []string{"profile", "sleep"},
		Endpoint:     fitbit.Endpoint,
		RedirectURL:  cfg.RedirectURL,
	}

	store := store.NewInMemoryStore()
	c := cron.New()

	setupCron(c, cfg.CronSchedule, oAuthConf, store)

	c.Start()

	http.HandleFunc("/", index)
	http.HandleFunc("/login", fitbitLogin(oAuthConf, store))
	http.HandleFunc("/fitbitCallback", fitbitCallback(oAuthConf, store))
	log.Println("Started running on http://127.0.0.1:8080")
	log.Println(http.ListenAndServe(":8080", nil))
}
