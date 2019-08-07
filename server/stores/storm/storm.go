package storm

import (
	"github.com/asdine/storm"
	"github.com/pkg/errors"
)

// Store TODO
type Store struct {
	db *storm.DB
}

// New TODO
func New(path string) (*Store, error) {
	db, err := storm.Open(path)
	if err != nil {
		return nil, errors.Wrapf(err, "could not open storm db at `%s`", path)
	}

	err = db.Init(&user{})
	err = db.Init(&node{})
	err = db.Init(&access{})
	if err != nil {
		return nil, errors.Wrapf(err, "could not initialize user storage")
	}

	return &Store{db}, nil
}
