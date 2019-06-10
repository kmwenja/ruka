package server

import "golang.org/x/crypto/ssh"

func handleNodeSession(backend Backend, c ssh.Conn, chans <-chan ssh.NewChannel, reqs <-chan *ssh.Request) {
}
