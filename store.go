package main

import (
	"sync"
)

// Store represents a store of Users
// Store must be safe for use by concurrent goroutines
type Store interface {
	Upsert(user User) error
	GetByState(state string) (*User, error)
	GetByEmail(email string) (*User, error)
	GetAll() ([]*User, error)
}

type inMemoryStore struct {
	users []*User
	mutex *sync.RWMutex
}

// NewInMemoryStore creates a new in memory user store
func NewInMemoryStore() Store {
	return &inMemoryStore{
		mutex: &sync.RWMutex{},
	}
}

func (i *inMemoryStore) Upsert(user User) error {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	for _, u := range i.users {
		if u.State == user.State {
			u.Email = user.Email
			u.Token = user.Token
			return nil
		}
	}

	i.users = append(i.users, &user)

	return nil
}

func (i *inMemoryStore) GetByState(state string) (*User, error) {
	i.mutex.RLock()
	defer i.mutex.RUnlock()

	for _, u := range i.users {
		if u.State == state {
			return u, nil
		}
	}

	return nil, nil
}

func (i *inMemoryStore) GetByEmail(email string) (*User, error) {
	i.mutex.RLock()
	defer i.mutex.RUnlock()

	for _, u := range i.users {
		if u.Email == email {
			return u, nil
		}
	}

	return nil, nil
}

func (i *inMemoryStore) GetAll() ([]*User, error) {
	return i.users, nil
}
