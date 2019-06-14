package control

import (
	"fmt"
	"io"
	"strings"

	"github.com/pkg/errors"
	"github.com/urfave/cli"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
)

// PROMPT TODO
const PROMPT string = ">>> "

// ShellBackend TODO
type ShellBackend interface {
	AddUser(username string, key []byte) error
}

// Shell TODO
type Shell struct {
	username  string
	b         ShellBackend
	t         *terminal.Terminal
	isRunning bool
}

// NewShell TODO
func NewShell(username string, backend ShellBackend, rw io.ReadWriter) *Shell {
	t := terminal.NewTerminal(rw, PROMPT)
	return &Shell{username, backend, t, false}
}

func (sh *Shell) writef(s string, args ...interface{}) error {
	_, err := sh.t.Write([]byte(fmt.Sprintf(s, args...)))
	return err
}

func (sh *Shell) readMulti() (string, error) {
	sh.t.SetPrompt("")
	defer sh.t.SetPrompt(PROMPT)

	l, err := sh.t.ReadLine()
	if err != nil {
		return "", err
	}

	s := l
	for l != "" {
		s += l
		l, err = sh.t.ReadLine()
		if err != nil {
			return "", err
		}
	}
	return s, nil
}

// Run TODO
func (sh *Shell) Run() error {
	app := cli.NewApp()
	app.Name = "control"
	app.Writer = sh.t
	app.ErrWriter = sh.t
	app.ExitErrHandler = func(c *cli.Context, err error) {
		if err != nil {
			return
		}

		if err == io.EOF {
			sh.isRunning = false
			return
		}

		sh.writef("Cli Error: %v\n", err)
	}

	app.Commands = []cli.Command{
		{
			Name:   "quit",
			Action: sh.quitCmd,
		},
		{
			Name:   "add_user",
			Action: sh.addUserCmd,
		},
		{
			Name:   "rm_user",
			Action: sh.notImplCmd,
		},
		{
			Name:   "add_user_key",
			Action: sh.notImplCmd,
		},
		{
			Name:   "rm_user_key",
			Action: sh.notImplCmd,
		},
	}

	sh.writef("Welcome to Ruka Control\n")
	sh.writef("Logged in as `%s`\n", sh.username)
	sh.writef("Type \"help\" for more info, Type \"quit\" to exit\n")
	sh.isRunning = true

	for sh.isRunning {
		l, err := sh.t.ReadLine()
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
			sh.writef("Cli Error: %v\n", err)
			continue
		}
	}

	return nil
}

func (sh *Shell) quitCmd(c *cli.Context) error {
	sh.isRunning = false
	return nil
}

func (sh *Shell) addUserCmd(c *cli.Context) error {
	username := c.Args().Get(0)
	key, err := sh.readMulti()
	if err != nil {
		return err
	}

	pk, _, _, _, err := ssh.ParseAuthorizedKey([]byte(key))
	if err != nil {
		return errors.Wrapf(err, "could not parse public key for `%s`", username)
	}

	err = sh.b.AddUser(username, pk.Marshal())
	if err != nil {
		return errors.Wrapf(err, "could not add user `%s`", username)
	}
	sh.writef("Added user `%s`\n", username)
	return nil
}

func (sh *Shell) notImplCmd(c *cli.Context) error {
	return fmt.Errorf("command not implemented")
}
