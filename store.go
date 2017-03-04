package main

// Store represents a store of Users
// Store must be safe for use by concurrent goroutines
type Store interface {
	Upsert(user User) error
	GetByState(state string) (*User, error)
	GetAll() ([]*User, error)
}
