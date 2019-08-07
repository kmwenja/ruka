package server

import "golang.org/x/crypto/ssh"

func handleJumpSession(s Store, c ssh.Conn, chans <-chan ssh.NewChannel, reqs <-chan *ssh.Request) {
}
