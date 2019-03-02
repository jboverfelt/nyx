package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/boltdb/bolt"
	mailgun "gopkg.in/mailgun/mailgun-go.v1"
	"github.com/robfig/cron"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/fitbit"
)

type config struct {
	ClientID          string `json:"clientId"`
	ClientSecret      string `json:"clientSecret"`
	RedirectURL       string `json:"redirectUrl"`
	CronSchedule      string `json:"cronSchedule"`
	MailgunPrivateKey string `json:"mailgunPrivateKey"`
	MailgunPublicKey  string `json:"mailgunPublicKey"`
	MailgunDomain     string `json:"mailgunDomain"`
}

func genConfig() config {
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

func main() {
	cfg := genConfig()

	oAuthConf := &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		Scopes:       []string{"profile", "sleep"},
		Endpoint:     fitbit.Endpoint,
		RedirectURL:  cfg.RedirectURL,
	}

	db, err := bolt.Open("users.db", 0600, &bolt.Options{Timeout: 5 * time.Second})

	if err != nil {
		log.Fatalf("Error opening database connection: %v\n", err)
	}

	store := NewBoltStore(db)

	env := &Env{
		OAuthConf: oAuthConf,
		DB:        store,
	}

	mg := mailgun.NewMailgun(cfg.MailgunDomain, cfg.MailgunPrivateKey, cfg.MailgunPublicKey)

	c := cron.New()

	setupCron(c, cfg.CronSchedule, oAuthConf, store, mg)

	c.Start()

	http.Handle("/", NewHandler(env, index))
	http.Handle("/login", NewHandler(env, fitbitLogin))
	http.Handle("/fitbitCallback", NewHandler(env, fitbitCallback))
	log.Println("Started running on http://127.0.0.1:8080")
	log.Println(http.ListenAndServe(":8080", nil))
}
