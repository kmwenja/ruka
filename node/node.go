package node

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"

	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
)

// Node TODO
type Node struct {
	cc *ssh.ClientConfig
}

// Config TODO
type Config struct {
	NodeKeyFile string
	HostKeyFile string
}

// New TODO
func New(cfg Config) (*Node, error) {
	nb, err := ioutil.ReadFile(cfg.NodeKeyFile)
	if err != nil {
		return nil, errors.Wrapf(err, "could not open node private key: %s", cfg.NodeKeyFile)
	}
	nk, err := ssh.ParsePrivateKey(nb)
	if err != nil {
		return nil, errors.Wrapf(err, "could not parse node private key: %s", cfg.NodeKeyFile)
	}

	hb, err := ioutil.ReadFile(cfg.HostKeyFile)
	if err != nil {
		return nil, errors.Wrapf(err, "could not open host public key: %s", cfg.HostKeyFile)
	}
	hk, _, _, _, err := ssh.ParseAuthorizedKey(hb)
	if err != nil {
		return nil, errors.Wrapf(err, "could not parse host public key: %s", cfg.HostKeyFile)
	}

	cc := &ssh.ClientConfig{
		User: "node",
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(nk),
		},
		HostKeyCallback: ssh.FixedHostKey(hk),
		ClientVersion:   "SSH-2.0-ruka-client-0.1", // TODO use actual version
	}
	return &Node{cc: cc}, nil
}

// Serve TODO
func (n *Node) Serve(addr string, c net.Conn) error {
	conn, chans, reqs, err := ssh.NewClientConn(c, addr, n.cc)
	if err != nil {
		return errors.Wrapf(err, "could not initiate ssh connection")
	}

	cli := ssh.NewClient(conn, chans, reqs)
	defer cli.Close()

	// handle node commands
	cmdChan := cli.HandleChannelOpen("ruka-node-commands")
	if cmdChan == nil {
		return errors.Wrapf(err, "`ruka-node-commands` handled by another goroutine")
	}
	go n.handleServerCommands(cmdChan)

	// setup reverse listen tunnel
	l, err := cli.Listen("tcp", "127.0.0.1:0") // choose port at random
	if err != nil {
		return errors.Wrapf(err, "could not start reverse tunnel")
	}
	defer l.Close()

	for {
		tc, err := l.Accept()
		if err != nil {
			return errors.Wrapf(err, "could not accept tunnel connection")
		}

		go func(tc net.Conn) {
			defer tc.Close()

			localSSH := "127.0.0.1:22"
			lc, err := net.Dial("tcp", localSSH)
			if err != nil {
				log.Printf("Error: could not connect to local SSH (%s): %v", localSSH, err)
				return
			}
			defer lc.Close()

			// redirect reads & writes
			go io.Copy(lc, tc)
			io.Copy(tc, lc)
		}(tc)
	}
}

func (n *Node) handleServerCommands(chans <-chan ssh.NewChannel) {
	for newChan := range chans {
		newChan.Reject(ssh.Prohibited, fmt.Sprintf("channel %s not allowed", newChan.ChannelType()))
	}
}
