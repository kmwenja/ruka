package storm

import (
	"fmt"
	"time"

	"github.com/asdine/storm"
	"github.com/kmwenja/ruka/server"
	"github.com/pkg/errors"
)

type user struct {
	Username string    `storm:"id,unique"`
	Key      []byte    `storm:"index"`
	Created  time.Time `storm:"index"`
	Updated  time.Time `storm:"index"`
}

// StoreUser TODO
func (s *Store) StoreUser(username string, key []byte) error {
	var u user
	tx, err := s.db.Begin(true)
	if err != nil {
		return errors.Wrapf(err, "could not start db transaction")
	}
	defer tx.Rollback()

	err = tx.One("Username", username, &u)
	if err != nil && err != storm.ErrNotFound {
		return errors.Wrapf(err, "could not query database")
	}

	if u.Username != "" {
		return fmt.Errorf("user already exists")
	}

	u.Username = username
	u.Key = key
	u.Created = time.Now()
	u.Updated = time.Now()
	err = tx.Save(&u)
	if err != nil {
		return errors.Wrapf(err, "could not save user")
	}

	err = tx.Commit()
	if err != nil {
		return errors.Wrapf(err, "could not commit db transaction")
	}

	return nil
}

// FetchUsers TODO
func (s *Store) FetchUsers() ([]server.User, error) {
	users := make([]server.User, 0)

	us := make([]user, 0)
	err := s.db.All(&us)
	if err != nil {
		if err == storm.ErrNotFound {
			return users, nil
		}
		return nil, errors.Wrapf(err, "could not query database")
	}

	for _, u := range us {
		usr := server.User{
			Username: u.Username,
			Created:  u.Created,
			Updated:  u.Updated,
		}
		users = append(users, usr)
	}

	return users, nil
}

// RemoveUser TODO
func (s *Store) RemoveUser(username string) error {
	var u user
	tx, err := s.db.Begin(true)
	if err != nil {
		return errors.Wrapf(err, "could not start db transaction")
	}
	defer tx.Rollback()

	err = tx.One("Username", username, &u)
	if err != nil {
		if err == storm.ErrNotFound {
			return errors.Wrap(err, "could not find user")
		}
		return errors.Wrapf(err, "could not query database")
	}
	err = tx.DeleteStruct(&u)
	if err != nil {
		return errors.Wrapf(err, "could not perform database delete")
	}

	err = tx.Commit()
	if err != nil {
		return errors.Wrapf(err, "could not commit transaction")
	}

	return nil
}
