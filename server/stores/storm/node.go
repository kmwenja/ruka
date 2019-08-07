package storm

import (
	"fmt"
	"time"

	"github.com/asdine/storm"
	"github.com/kmwenja/ruka/server"
	"github.com/pkg/errors"
)

type node struct {
	Name    string    `storm:"id,unique"`
	Key     []byte    `storm:"index"`
	Created time.Time `storm:"index"`
	Updated time.Time `storm:"index"`
}

// StoreNode TODO
func (s *Store) StoreNode(name string, key []byte) error {
	var n node
	tx, err := s.db.Begin(true)
	if err != nil {
		return errors.Wrapf(err, "could not start db transaction")
	}
	defer tx.Rollback()

	err = tx.One("Name", name, &n)
	if err != nil && err != storm.ErrNotFound {
		return errors.Wrapf(err, "could not query database")
	}

	if n.Name != "" {
		return fmt.Errorf("node already exists")
	}

	n.Name = name
	n.Key = key
	n.Created = time.Now()
	n.Updated = time.Now()
	err = tx.Save(&n)
	if err != nil {
		return errors.Wrapf(err, "could not save node")
	}

	err = tx.Commit()
	if err != nil {
		return errors.Wrapf(err, "could not commit db transaction")
	}

	return nil
}

// FetchNodes TODO
func (s *Store) FetchNodes() ([]server.Node, error) {
	nodes := make([]server.Node, 0)

	nds := make([]node, 0)
	err := s.db.All(&nds)
	if err != nil {
		if err == storm.ErrNotFound {
			return nodes, nil
		}
		return nil, errors.Wrapf(err, "could not query database")
	}

	for _, n := range nds {
		nd := server.Node{
			Name:    n.Name,
			Created: n.Created,
			Updated: n.Updated,
		}
		nodes = append(nodes, nd)
	}

	return nodes, nil
}

// RemoveNode TODO
func (s *Store) RemoveNode(name string) error {
	var n node
	tx, err := s.db.Begin(true)
	if err != nil {
		return errors.Wrapf(err, "could not start db transaction")
	}
	defer tx.Rollback()

	err = tx.One("Name", name, &n)
	if err != nil {
		if err == storm.ErrNotFound {
			return errors.Wrap(err, "could not find node")
		}
		return errors.Wrapf(err, "could not query database")
	}
	err = tx.DeleteStruct(&n)
	if err != nil {
		return errors.Wrapf(err, "could not perform database delete")
	}

	err = tx.Commit()
	if err != nil {
		return errors.Wrapf(err, "could not commit transaction")
	}

	return nil
}
