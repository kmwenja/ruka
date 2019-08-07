package server

import (
	"time"

	"github.com/pkg/errors"
	"github.com/urfave/cli"
	"golang.org/x/crypto/ssh"
)

// User TODO
type User struct {
	Username string
	Created  time.Time
	Updated  time.Time
}

func (s *Server) addUserCmd(t *sshTerminal) cli.Command {
	return cli.Command{
		Name: "add_user",
		Action: func(c *cli.Context) error {
			username := c.Args().Get(0)
			key, err := t.ReadLines()
			if err != nil {
				return errors.Wrapf(err, "could not read key from terminal")
			}

			pk, _, _, _, err := ssh.ParseAuthorizedKey([]byte(key))
			if err != nil {
				return errors.Wrapf(err, "could not parse public key")
			}
			err = s.s.StoreUser(username, pk.Marshal())
			if err != nil {
				return errors.Wrapf(err, "could not add user")
			}

			return t.Printf("Added user `%s`\n", username)
		},
	}
}

func (s *Server) lsUsersCmd(t *sshTerminal) cli.Command {
	return cli.Command{
		Name: "ls_users",
		Action: func(c *cli.Context) error {
			users, err := s.s.FetchUsers()
			if err != nil {
				return errors.Wrapf(err, "could not fetch users")
			}

			for _, u := range users {
				t.Printf(
					"%s\t%s\t%s\n", u.Username, u.Created, u.Updated)
			}

			return nil
		},
	}
}

func (s *Server) rmUserCmd(t *sshTerminal) cli.Command {
	return cli.Command{
		Name: "rm_user",
		Action: func(c *cli.Context) error {
			username := c.Args().Get(0)
			err := s.s.RemoveUser(username)
			if err != nil {
				return errors.Wrapf(err, "could not remove user `%s`", username)
			}
			return t.Printf("Removed user `%s`\n", username)
		},
	}
}
