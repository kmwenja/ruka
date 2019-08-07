package storm

import (
	"time"

	"github.com/asdine/storm"
	"github.com/kmwenja/ruka/server"
	"github.com/pkg/errors"
)

// AccessID TODO
type AccessID struct {
	Username string
	Node     string
}

type access struct {
	ID      AccessID  `storm:"id,unique"`
	Key     []byte    `storm:"index"`
	Created time.Time `storm:"index"`
}

// TODO expiry

// StoreAccessRecord TODO
func (s *Store) StoreAccessRecord(username, nodename string, key []byte) error {
	var u user
	var n node
	var a access

	tx, err := s.db.Begin(true)
	if err != nil {
		return errors.Wrapf(err, "could not start db transaction")
	}
	defer tx.Rollback()

	err = tx.One("Username", username, &u)
	if err != nil {
		if err == storm.ErrNotFound {
			return errors.Wrapf(err, "could not find user: %s", username)
		}
		return errors.Wrapf(err, "could not query database")
	}

	err = tx.One("Name", nodename, &n)
	if err != nil {
		if err == storm.ErrNotFound {
			return errors.Wrapf(err, "could not find node: %s", nodename)
		}
		return errors.Wrapf(err, "could not query database")
	}

	id := AccessID{username, nodename}
	err = tx.One("ID", id, &a)
	if err == storm.ErrNotFound {
		return errors.Wrapf(err, "access for `%s` to `%s already exists", username, nodename)
	}
	if err != nil {
		return errors.Wrapf(err, "could not query database")
	}

	a.ID = id
	a.Key = key
	a.Created = time.Now()
	err = tx.Save(&a)
	if err != nil {
		return errors.Wrapf(err, "could not save access")
	}

	err = tx.Commit()
	if err != nil {
		return errors.Wrapf(err, "could not commit db transaction")
	}

	return nil
}

// FetchAccessRecords TODO
func (s *Store) FetchAccessRecords() ([]server.AccessRecord, error) {
	records := make([]server.AccessRecord, 0)

	ars := make([]access, 0)
	err := s.db.All(&ars)
	if err != nil {
		if err == storm.ErrNotFound {
			return records, nil
		}
		return nil, errors.Wrapf(err, "could not query database")
	}

	for _, ar := range ars {
		record := server.AccessRecord{
			Username:  ar.ID.Username,
			Node:      ar.ID.Node,
			Timestamp: ar.Created,
		}
		records = append(records, record)
	}

	return records, nil
}

// RemoveAccessRecord TODO
func (s *Store) RemoveAccessRecord(username, nodename string) error {
	var a access
	tx, err := s.db.Begin(true)
	if err != nil {
		return errors.Wrapf(err, "could not start db transaction")
	}
	defer tx.Rollback()

	id := AccessID{username, nodename}
	err = tx.One("ID", id, &a)
	if err != nil {
		if err == storm.ErrNotFound {
			return errors.Wrap(err, "could not find access record")
		}
		return errors.Wrapf(err, "could not query database")
	}
	err = tx.DeleteStruct(&a)
	if err != nil {
		return errors.Wrapf(err, "could not perform database delete")
	}

	err = tx.Commit()
	if err != nil {
		return errors.Wrapf(err, "could not commit transaction")
	}

	return nil
}
