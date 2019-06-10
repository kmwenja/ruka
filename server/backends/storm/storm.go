package storm

import (
	"fmt"

	storm "github.com/asdine/storm"
	"github.com/kmwenja/ruka/server"
	"github.com/pkg/errors"
)

// Backend TODO
type Backend struct {
	db *storm.DB
}

type user struct {
	ID       int    `storm:"id,increment"`
	Username string `storm:"unique"`
	Key      []byte `storm:"index"`
}

// New TODO
func New(path string) (*Backend, error) {
	db, err := storm.Open(path)
	if err != nil {
		return nil, errors.Wrapf(err, "could not open storm db at `%s`", path)
	}

	err = db.Init(&user{})
	if err != nil {
		return nil, errors.Wrapf(err, "could not initialize user storage")
	}

	return &Backend{db}, nil
}

// Authenticate TODO
func (b *Backend) Authenticate(st server.SessionType, key []byte) (string, error) {
	switch st {
	case server.SessionType_CONTROL:
		var u user
		err := b.db.One("Key", key, &u)
		if err != nil {
			if err == storm.ErrNotFound {
				return "", errors.Wrapf(err, "key not found")
			}
			return "", errors.Wrapf(err, "could not query database")
		}

		return string(u.ID), nil
	default:
		return "", fmt.Errorf("mode not supported: %s", st)
	}

}

// AddUser TODO
func (b *Backend) AddUser(username string, key []byte) error {
	var u user
	err := b.db.One("Username", username, &u)
	if err != nil && err != storm.ErrNotFound {
		return errors.Wrapf(err, "could not query database")
	}
	if u.Username != "" {
		return nil
	}

	u.Username = username
	u.Key = key
	err = b.db.Save(&u)
	if err != nil {
		return errors.Wrapf(err, "could not save user")
	}
	return nil
}
