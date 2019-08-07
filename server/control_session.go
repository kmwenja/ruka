package server

import (
	"context"
	"encoding/binary"
	"log"

	"golang.org/x/crypto/ssh"
)

func (s *Server) handleControlSession(username string, chans <-chan ssh.NewChannel) {
	log.Printf("Control Session: Start")

	for nC := range chans {
		if nC.ChannelType() != "session" {
			nC.Reject(ssh.Prohibited, "unsupported channel type")
			continue
		}

		channel, requests, err := nC.Accept()
		if err != nil {
			log.Printf("Control Session: Error: could not accept channel: %v", err)
			continue
		}

		go s.handleControlChannel(username, channel, requests)
	}
	log.Printf("Control Session: End")
}

func (s *Server) handleControlChannel(username string, channel ssh.Channel, reqs <-chan *ssh.Request) {
	defer channel.Close()

	// TODO handle window size changes

	ctx := context.TODO()

OUTER:
	for req := range reqs {
		switch req.Type {
		case "shell":
			// only on the shell request do we actually do something
			st := newSSHTerminal(channel, ">>> ")
			err := s.shell(ctx, st, username)
			if err != nil {
				log.Printf("Control Session: Error: shell error: %v", err)
			}
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
