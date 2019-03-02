package main

import (
	"fmt"
	"html/template"
	"log"
	"time"

	"bytes"

	mailgun "gopkg.in/mailgun/mailgun-go.v1"
)

func genNoSleepMessage() string {
	return `
You have no sleep records from last night.
Were you wearing your Fitbit?
    `
}

func genSleepMessage(sleep SleepResponse) (string, error) {
	t, err := time.Parse("2006-01-02T15:04:05.999999999", sleep.Sleep[0].StartTime)

	if err != nil {
		return "", err
	}

	dur, err := time.ParseDuration(fmt.Sprintf("%dm", sleep.Summary.TotalMinutesAsleep))

	if err != nil {
		return "", err
	}

	data := struct {
		Efficiency       int
		StartTime        string
		TotalHoursAsleep string
		AwakeCount       int
		AwakeDuration    int
		RestlessCount    int
		RestlessDuration int
	}{
		Efficiency:       sleep.Sleep[0].Efficiency,
		StartTime:        t.Format(time.Kitchen),
		TotalHoursAsleep: fmt.Sprintf("%.2f", dur.Hours()),
		AwakeCount:       sleep.Sleep[0].AwakeCount,
		AwakeDuration:    sleep.Sleep[0].AwakeDuration,
		RestlessCount:    sleep.Sleep[0].RestlessCount,
		RestlessDuration: sleep.Sleep[0].RestlessDuration,
	}

	tmpl := template.Must(template.ParseFiles("templates/email.tmpl"))

	var b bytes.Buffer

	err = tmpl.Execute(&b, data)

	if err != nil {
		return "", err
	}

	return b.String(), nil
}

func sendEmail(mg mailgun.Mailgun, sleep SleepResponse, u *User) error {
	fromAddr := "nyx@" + mg.Domain()
	subject := "Last Night's Sleep"
	toAddr := u.Email
	var body string

	if sleep.Summary.TotalSleepRecords > 0 {
		var err error
		body, err = genSleepMessage(sleep)

		if err != nil {
			return err
		}
	} else {
		body = genNoSleepMessage()
	}

	log.Printf("Sending message to %s\n", u.Email)
	msg := mg.NewMessage(fromAddr, subject, body, toAddr)
	_, _, err := mg.Send(msg)

	if err != nil {
		return err
	}

	return nil
}
