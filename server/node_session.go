package server

import "golang.org/x/crypto/ssh"

func handleNodeSession(s Store, c ssh.Conn, chans <-chan ssh.NewChannel, reqs <-chan *ssh.Request) {
	// establish tunnel
	// register node with node manager
}
