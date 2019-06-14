package server

import (
	"encoding/binary"
	"log"

	"github.com/kmwenja/ruka/server/control"
	"golang.org/x/crypto/ssh"
)

func handleControlSession(backend Backend, c *ssh.ServerConn, chans <-chan ssh.NewChannel, reqs <-chan *ssh.Request) {
	log.Printf("Control Session: Start")
	go ssh.DiscardRequests(reqs)

	username := c.Permissions.Extensions["ruka-username"]

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

		go handleControlChannel(username, backend, channel, requests)
	}
	log.Printf("Control Session: End")
}

func handleControlChannel(username string, backend Backend, channel ssh.Channel, reqs <-chan *ssh.Request) {
	defer channel.Close()

	// TODO handle window size changes

OUTER:
	for req := range reqs {
		switch req.Type {
		case "shell":
			// only on the shell request do we actually do something
			sh := control.NewShell(username, backend, channel)
			err := sh.Run()
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
