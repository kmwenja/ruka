package server

import (
	"fmt"
	"io"
	"strings"

	"github.com/pkg/errors"
	"github.com/urfave/cli"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
)

// ControlShell TODO
func ControlShell(backend Backend, rw io.ReadWriter) {
	t := terminal.NewTerminal(rw, ">>> ")
	write := func(s string) error {
		_, err := t.Write([]byte(s))
		return err
	}
	writef := func(s string, args ...interface{}) error {
		return write(fmt.Sprintf(s, args...))
	}
	readMulti := func() (string, error) {
		t.SetPrompt("")
		defer t.SetPrompt(">>> ")

		l, err := t.ReadLine()
		if err != nil {
			return "", err
		}

		s := l
		for l != "" {
			s += l
			l, err = t.ReadLine()
			if err != nil {
				return "", err
			}
		}
		return s, nil
	}

	loop := true
	app := cli.NewApp()
	app.Name = "control"
	app.Writer = t
	app.ErrWriter = t
	app.ExitErrHandler = func(c *cli.Context, err error) {
		if err != nil {
			return
		}

		if err == io.EOF {
			loop = false
			return
		}

		writef("Cli Error: %v\n", err)
	}

	app.Commands = []cli.Command{
		{
			Name: "quit",
			Action: func(c *cli.Context) error {
				loop = false
				return nil
			},
		},
		{
			Name: "add_user",
			Action: func(c *cli.Context) error {
				username := c.Args().Get(0)
				key, err := readMulti()
				if err != nil {
					return err
				}

				pk, _, _, _, err := ssh.ParseAuthorizedKey([]byte(key))
				if err != nil {
					return errors.Wrapf(err, "could not parse public key for `%s`", username)
				}

				err = backend.AddUser(username, pk.Marshal())
				if err != nil {
					return errors.Wrapf(err, "could not add user `%s`", username)
				}
				writef("Added user `%s`\n", username)
				return nil
			},
		},
	}

	write("Welcome to Ruka Control\nType \"help\" for more info, Type \"quit\" to exit\n")
	for loop {
		l, err := t.ReadLine()
		if err == io.EOF {
			return
		}

		if err != nil {
			writef("Error: %v\n", err)
			return
		}

		args := strings.Fields(l)
		args = append([]string{"control"}, args...)
		err = app.Run(args)
		if err != nil {
			writef("Cli Error: %v\n", err)
			continue
		}
	}
}
