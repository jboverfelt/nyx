package main

import (
	"encoding/json"

	"github.com/boltdb/bolt"
)

const userBucket = "users"

type boltStore struct {
	db *bolt.DB
}

// NewBoltStore creates a new implementation
// of the Store interface which uses BoltDB
func NewBoltStore(d *bolt.DB) Store {
	return &boltStore{
		db: d,
	}
}

func (b *boltStore) Upsert(u User) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(userBucket))

		if err != nil {
			return err
		}

		jsonStr, err := json.Marshal(u)

		if err != nil {
			return err
		}

		err = bucket.Put([]byte(u.State), jsonStr)

		if err != nil {
			return err
		}

		return nil
	})
}

func (b *boltStore) GetByState(state string) (*User, error) {
	tx, err := b.db.Begin(false)

	if err != nil {
		return nil, err
	}

	// read only transactions are always rolled back
	defer tx.Rollback()

	bucket := tx.Bucket([]byte(userBucket))

	if bucket == nil {
		return nil, nil
	}

	val := bucket.Get([]byte(state))

	if val == nil {
		return nil, nil
	}

	var u User
	err = json.Unmarshal(val, &u)

	if err != nil {
		return nil, err
	}

	return &u, nil
}

func (b *boltStore) GetAll() ([]*User, error) {
	var users []*User

	err := b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(userBucket))

		// if bucket hasn't been created, then there are no users
		if bucket == nil {
			return nil
		}

		err := bucket.ForEach(func(k, v []byte) error {
			var u User

			err := json.Unmarshal(v, &u)

			if err != nil {
				return err
			}

			users = append(users, &u)

			return nil
		})

		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return users, nil
}
