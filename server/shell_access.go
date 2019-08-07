package server

import (
	"crypto/rand"
	"crypto/rsa"
	"time"

	"github.com/pkg/errors"
	"github.com/urfave/cli"
	"golang.org/x/crypto/ssh"
)

// AccessRecord TODO
type AccessRecord struct {
	Username  string
	Node      string
	Timestamp time.Time
}

func (s *Server) lsAccessCmd(t *sshTerminal) cli.Command {
	return cli.Command{
		Name: "ls_access",
		Action: func(c *cli.Context) error {
			as, err := s.s.FetchAccessRecords()
			if err != nil {
				return errors.Wrapf(err, "could not fetch access records")
			}

			for _, a := range as {
				t.Printf("%s\t%s\t%s\n", a.Username, a.Node, a.Timestamp)
			}

			return nil
		},
	}
}

func (s *Server) allowAccessCmd(t *sshTerminal) cli.Command {
	return cli.Command{
		Name: "allow_access",
		Action: func(c *cli.Context) error {
			username := c.Args().Get(0)
			nodename := c.Args().Get(1)

			rk, err := rsa.GenerateKey(rand.Reader, 2048)
			if err != nil {
				return errors.Wrapf(err, "could not generate rsa access key")
			}

			key, err := ssh.NewSignerFromKey(rk)
			if err != nil {
				return errors.Wrapf(err, "could not make ssh access key from rsa key")
			}

			err = s.s.StoreAccessRecord(username, nodename, ssh.Marshal(key))
			if err != nil {
				return errors.Wrapf(err, "could not allow access")
			}

			return t.Printf("Allow `%s` to access `%s`\n", username, nodename)
		},
	}
}

func (s *Server) denyAccessCmd(t *sshTerminal) cli.Command {
	return cli.Command{
		Name: "deny_access",
		Action: func(c *cli.Context) error {
			username := c.Args().Get(0)
			nodename := c.Args().Get(1)
			err := s.s.RemoveAccessRecord(username, nodename)
			if err != nil {
				return errors.Wrapf(err, "could not deny access")
			}

			return t.Printf("Deny `%s` to access `%s`\n", username, nodename)
		},
	}
}
