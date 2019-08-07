package server

import (
	"context"
	"io"
	"strings"

	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func (s *Server) shell(ctx context.Context, t *sshTerminal, username string) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	app := cli.NewApp()
	app.Name = "control"
	app.Writer = t
	app.ErrWriter = t
	app.ExitErrHandler = func(c *cli.Context, err error) {
		if err != nil {
			return
		}

		if err == io.EOF {
			cancel()
			return
		}

		t.Printf("Cli Error: %v\n", err)
	}

	app.Commands = []cli.Command{
		cli.Command{
			Name: "quit",
			Action: func(c *cli.Context) error {
				cancel()
				return nil
			},
		},

		s.lsUsersCmd(t),
		s.addUserCmd(t),
		s.rmUserCmd(t),

		s.lsNodesCmd(t),
		s.addNodeCmd(t),
		s.rmNodeCmd(t),

		s.lsAccessCmd(t),
		s.allowAccessCmd(t),
		s.denyAccessCmd(t),
	}

	t.Printf("Welcome to Ruka Control\n")
	t.Printf("Logged in as `%s`\n", username)
	t.Printf("Type \"help\" for more info, Type \"quit\" to exit\n")

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			l, err := t.ReadLine()
			if err == io.EOF {
				return nil
			}

			if err != nil {
				return errors.Wrapf(err, "could not read from terminal")
			}

			args := strings.Fields(l)
			args = append([]string{"control"}, args...)
			err = app.Run(args)
			if err != nil {
				t.Printf("Cli Error: %v\n", err)
				continue
			}
		}
	}
}
