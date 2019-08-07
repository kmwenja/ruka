package storm

import (
	"github.com/asdine/storm"
	"github.com/pkg/errors"
)

// FetchUsernameByKey TODO
func (s *Store) FetchUsernameByKey(key []byte) (string, error) {
	var u user
	err := s.fetchWithKey(key, &u)
	return u.Username, err
}

// FetchNodenameByKey TODO
func (s *Store) FetchNodenameByKey(key []byte) (string, error) {
	var n node
	err := s.fetchWithKey(key, &n)
	return n.Name, err
}

func (s *Store) fetchWithKey(key []byte, i interface{}) error {
	err := s.db.One("Key", key, i)
	if err != nil {
		if err == storm.ErrNotFound {
			return errors.Wrapf(err, "key not found")
		}
		return errors.Wrapf(err, "could not query database")
	}
	return nil
}

// FetchKeyForJump TODO
func (s *Store) FetchKeyForJump(username string, node string) ([]byte, error) {
	var a access
	id := AccessID{username, node}
	err := s.db.One("ID", id, &a)
	if err != nil {
		if err == storm.ErrNotFound {
			return nil, errors.Wrapf(err, "access not found")
		}
		return nil, errors.Wrapf(err, "could not query database")
	}
	return a.Key, nil
}
