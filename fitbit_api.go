package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/mailgun/mailgun-go"
	"github.com/robfig/cron"
	"golang.org/x/oauth2"
)

// SleepResponse represents a response from the Fitbit API
// for sleep
type SleepResponse struct {
	Sleep []struct {
		AwakeCount      int    `json:"awakeCount"`
		AwakeDuration   int    `json:"awakeDuration"`
		AwakeningsCount int    `json:"awakeningsCount"`
		DateOfSleep     string `json:"dateOfSleep"`
		Duration        int    `json:"duration"`
		Efficiency      int    `json:"efficiency"`
		IsMainSleep     bool   `json:"isMainSleep"`
		LogID           int64  `json:"logId"`
		MinuteData      []struct {
			DateTime string `json:"dateTime"`
			Value    string `json:"value"`
		} `json:"minuteData"`
		MinutesAfterWakeup  int    `json:"minutesAfterWakeup"`
		MinutesAsleep       int    `json:"minutesAsleep"`
		MinutesAwake        int    `json:"minutesAwake"`
		MinutesToFallAsleep int    `json:"minutesToFallAsleep"`
		RestlessCount       int    `json:"restlessCount"`
		RestlessDuration    int    `json:"restlessDuration"`
		StartTime           string `json:"startTime"`
		TimeInBed           int    `json:"timeInBed"`
	} `json:"sleep"`
	Summary struct {
		TotalMinutesAsleep int `json:"totalMinutesAsleep"`
		TotalSleepRecords  int `json:"totalSleepRecords"`
		TotalTimeInBed     int `json:"totalTimeInBed"`
	} `json:"summary"`
}

// Taken from rakyll's implementation in issue 84
// https://github.com/golang/oauth2/issues/84#issuecomment-175834679
type cacherTransport struct {
	Base *oauth2.Transport
	s    Store
}

// fitbit api refresh tokens are single use, so must save new ones
func (c *cacherTransport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	tok, err := c.Base.Source.Token()
	if err != nil {
		return nil, err
	}

	resp, err = c.Base.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	newTok, err := c.Base.Source.Token()
	if err != nil {
		return nil, err
	}

	if tok.AccessToken != newTok.AccessToken {
		err = c.s.UpdateByAccessToken(tok, newTok)

		if err != nil {
			return nil, err
		}
	}

	return resp, nil
}

func genFitbitURL() string {
	urlFirst := "https://api.fitbit.com/1/user/-/sleep/date/"
	urlSecond := ".json?isMainSleep=true"

	curDate := time.Now().Format("2006-01-02")

	url := urlFirst + curDate + urlSecond

	return url
}

func parseToken(tok string) (*oauth2.Token, error) {
	var token oauth2.Token
	err := json.Unmarshal([]byte(tok), &token)

	if err != nil {
		return nil, err
	}

	return &token, nil
}

func getSleepData(oAuthConf *oauth2.Config, token *oauth2.Token, s Store) (*SleepResponse, error) {
	ts := oAuthConf.TokenSource(oauth2.NoContext, token)
	tr := &oauth2.Transport{Source: ts}

	client := &http.Client{
		Transport: &cacherTransport{Base: tr, s: s},
	}

	resp, err := client.Get(genFitbitURL())

	if err != nil {
		return nil, fmt.Errorf("ERROR: fitbit call failed: %v\n", err.Error())
	}

	defer resp.Body.Close()

	var sleepResponse SleepResponse

	err = json.NewDecoder(resp.Body).Decode(&sleepResponse)

	if err != nil {
		return nil, fmt.Errorf("ERROR: failed to decode API response: %v", err)
	}

	return &sleepResponse, nil
}

func sleepChecker(oAuthConf *oauth2.Config, s Store, mg mailgun.Mailgun) func() {
	return func() {
		log.Println("Starting up cron")
		us, err := s.GetAll()

		if err != nil {
			log.Printf("ERROR: Fetching all users failed: %v\n", err)
		}

		if us == nil {
			log.Println("No users found, quitting cron")
			return
		}

		for _, u := range us {
			if u.Token == "" {
				log.Printf("ERROR: token was nil for email %v\n", u.Email)
				continue
			}

			token, err := parseToken(u.Token)

			if err != nil {
				log.Printf("ERROR: failed to unmarshal token for email %v\n", u.Email)
				continue
			}

			sleep, err := getSleepData(oAuthConf, token, s)

			if err != nil {
				log.Printf("ERROR: failed to get sleep data: %v\n", err)
				continue
			}

			err = sendEmail(mg, *sleep, u)

			if err != nil {
				log.Printf("ERROR: failed to send email: %v\n", err)
				continue
			}
		}
	}
}

func setupCron(c *cron.Cron, schedule string, oAuthConf *oauth2.Config, s Store, mg mailgun.Mailgun) {
	log.Printf("Scheduling email func on schedule: %s\n", schedule)

	err := c.AddFunc(schedule, sleepChecker(oAuthConf, s, mg))

	if err != nil {
		log.Fatalf("Could not schedule cron func %v\n", err)
	}
}
