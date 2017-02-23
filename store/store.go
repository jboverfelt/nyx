package store

import "github.com/jboverfelt/nyx/models"

// Store represents a store of Users
type Store interface {
	Upsert(user models.User) error
	GetByState(state string) (*models.User, error)
	GetByEmail(email string) (*models.User, error)
}

type inMemoryStore struct {
	users []*models.User
}

// NewInMemoryStore creates a new in memory user store
func NewInMemoryStore() Store {
	return &inMemoryStore{}
}

func (i *inMemoryStore) Upsert(user models.User) error {
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

func (i *inMemoryStore) GetByState(state string) (*models.User, error) {
	for _, u := range i.users {
		if u.State == state {
			return u, nil
		}
	}

	return nil, nil
}

func (i *inMemoryStore) GetByEmail(email string) (*models.User, error) {
	for _, u := range i.users {
		if u.Email == email {
			return u, nil
		}
	}

	return nil, nil
}
