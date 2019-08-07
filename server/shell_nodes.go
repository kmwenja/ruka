package server

import (
	"time"

	"github.com/pkg/errors"
	"github.com/urfave/cli"
	"golang.org/x/crypto/ssh"
)

// Node TODO
type Node struct {
	Name    string
	Created time.Time
	Updated time.Time
}

func (s *Server) lsNodesCmd(t *sshTerminal) cli.Command {
	return cli.Command{
		Name: "ls_nodes",
		Action: func(c *cli.Context) error {
			nodes, err := s.s.FetchNodes()
			if err != nil {
				return errors.Wrapf(err, "could not fetch nodes")
			}

			for _, n := range nodes {
				t.Printf("%s\t%s\t%s\n", n.Name, n.Created, n.Updated)
			}

			return nil
		},
	}
}

func (s *Server) addNodeCmd(t *sshTerminal) cli.Command {
	return cli.Command{
		Name: "add_node",
		Action: func(c *cli.Context) error {
			name := c.Args().Get(0)
			key, err := t.ReadLines()
			if err != nil {
				return errors.Wrapf(err, "could not read key from terminal")
			}

			pk, _, _, _, err := ssh.ParseAuthorizedKey([]byte(key))
			if err != nil {
				return errors.Wrapf(err, "could not parse public key")
			}

			err = s.s.StoreNode(name, pk.Marshal())
			if err != nil {
				return errors.Wrapf(err, "could not add node")
			}

			return t.Printf("Added node `%s`\n", name)
		},
	}
}

func (s *Server) rmNodeCmd(t *sshTerminal) cli.Command {
	return cli.Command{
		Name: "rm_node",
		Action: func(c *cli.Context) error {
			name := c.Args().Get(0)
			err := s.s.RemoveNode(name)
			if err != nil {
				return errors.Wrapf(err, "could not remove node")
			}
			return t.Printf("Removed node `%s`\n", name)
		},
	}
}
