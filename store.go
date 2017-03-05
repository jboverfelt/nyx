package main

import (
	"golang.org/x/oauth2"
)

// Store represents a store of Users
// Store must be safe for use by concurrent goroutines
type Store interface {
	Upsert(user User) error
	UpdateByAccessToken(oldToken, newToken *oauth2.Token) error
	GetByState(state string) (*User, error)
	GetAll() ([]*User, error)
}
