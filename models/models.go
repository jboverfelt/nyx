package models

// User represents a user of the application
type User struct {
	// id
	State string
	Email string
	Token string
}

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
