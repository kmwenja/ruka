package server

import (
	"fmt"
	"io"

	sshterm "golang.org/x/crypto/ssh/terminal"
)

type sshTerminal struct {
	t      *sshterm.Terminal
	prompt string
}

func newSSHTerminal(rw io.ReadWriter, prompt string) *sshTerminal {
	t := sshterm.NewTerminal(rw, prompt)
	return &sshTerminal{t, prompt}
}

func (s *sshTerminal) ReadLine() (string, error) {
	return s.t.ReadLine()
}

func (s *sshTerminal) ReadLines() (string, error) {
	s.t.SetPrompt("")
	defer s.t.SetPrompt(s.prompt)

	l, err := s.t.ReadLine()
	if err != nil {
		return "", err
	}

	res := l
	for l != "" {
		res += l
		l, err = s.t.ReadLine()
		if err != nil {
			return "", err
		}
	}
	return res, nil
}

func (s *sshTerminal) Printf(format string, args ...interface{}) error {
	_, err := s.Write([]byte(fmt.Sprintf(format, args...)))
	return err
}

func (s *sshTerminal) Write(p []byte) (int, error) {
	return s.t.Write(p)
}
