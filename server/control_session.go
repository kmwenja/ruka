package server

import (
	"encoding/binary"
	"log"

	"golang.org/x/crypto/ssh"
)

func handleControlSession(backend Backend, c ssh.Conn, chans <-chan ssh.NewChannel, reqs <-chan *ssh.Request) {
	log.Printf("Control Session: Start")
	go ssh.DiscardRequests(reqs)

	for nC := range chans {
		if nC.ChannelType() != "session" {
			nC.Reject(ssh.Prohibited, "unsupported channel type")
			continue
		}

		channel, requests, err := nC.Accept()
		if err != nil {
			log.Printf("Error: could not accept channel: %v", err)
			continue
		}

		go handleControlChannel(backend, channel, requests)
	}
}

func handleControlChannel(backend Backend, channel ssh.Channel, reqs <-chan *ssh.Request) {
	defer channel.Close()

	// TODO handle window size changes

OUTER:
	for req := range reqs {
		switch req.Type {
		case "shell":
			// only on the shell request do we actually do something
			ControlShell(backend, channel)
			break OUTER
		default:
			// ignore every request but accept them anyway to keep it moving
			req.Reply(true, nil)
		}
	}

	exitStatus := make([]byte, 4)
	binary.BigEndian.PutUint32(exitStatus, 0)
	_, err := channel.SendRequest("exit-status", false, exitStatus)
	if err != nil {
		log.Printf("Control Session: Error: could not send exit-status: %v", err)
	}
}
