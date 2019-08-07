package server

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"

	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
)

// Store TODO
type Store interface {
	StoreUser(username string, key []byte) error
	FetchUsers() ([]User, error)
	RemoveUser(username string) error

	StoreNode(name string, key []byte) error
	FetchNodes() ([]Node, error)
	RemoveNode(name string) error

	StoreAccessRecord(username string, node string, key []byte) error
	FetchAccessRecords() ([]AccessRecord, error)
	RemoveAccessRecord(username string, node string) error

	FetchUsernameByKey(key []byte) (string, error)
	FetchNodenameByKey(key []byte) (string, error)
	FetchKeyForJump(username string, nodename string) ([]byte, error)
}

// Server TODO
type Server struct {
	s       Store
	rootKey ssh.PublicKey
	hostKey ssh.Signer
}

// Config TODO
type Config struct {
	Store       Store
	HostKeyFile string
	RootKeyFile string
}

// New TODO
func New(cfg Config) (*Server, error) {
	hb, err := ioutil.ReadFile(cfg.HostKeyFile)
	if err != nil {
		return nil, errors.Wrapf(err, "could not open host private key: %s", cfg.HostKeyFile)
	}
	hk, err := ssh.ParsePrivateKey(hb)
	if err != nil {
		return nil, errors.Wrapf(err, "could not parse host private key: %s", cfg.HostKeyFile)
	}

	rb, err := ioutil.ReadFile(cfg.RootKeyFile)
	if err != nil {
		return nil, errors.Wrapf(err, "could not open root public key: %s", cfg.RootKeyFile)
	}
	rk, _, _, _, err := ssh.ParseAuthorizedKey(rb)
	if err != nil {
		return nil, errors.Wrapf(err, "could not parse root public key: %s", cfg.RootKeyFile)
	}

	return &Server{cfg.Store, rk, hk}, nil
}

// Serve TODO
func (s *Server) Serve(l net.Listener) error {
	for {
		c, err := l.Accept()
		if err != nil {
			return errors.Wrapf(err, "could not accept connection")
		}

		go func(c net.Conn) {
			err := s.ServeConnection(c)
			if err != nil {
				log.Printf("Error: could not handle connection: %v", err)
			}
		}(c)
	}
}

// ServeConnection TODO
func (s *Server) ServeConnection(c net.Conn) error {
	defer c.Close()

	sc := &ssh.ServerConfig{
		MaxAuthTries:      3, // TODO customize
		PublicKeyCallback: s.authenticatePubKey(),
		AuthLogCallback:   nil,                       // TODO log auths
		ServerVersion:     "SSH-2.0-ruka-server-0.1", // TODO use actual version
		BannerCallback:    nil,                       // TODO nice banner
	}
	sc.AddHostKey(s.hostKey)

	conn, chans, reqs, err := ssh.NewServerConn(c, sc)
	if err != nil {
		return errors.Wrapf(err, "could not initialize ssh connection")
	}
	defer conn.Close()

	st := conn.Permissions.Extensions["ruka-session-type"]
	switch st {
	case "node":
		handleNodeSession(s.s, conn, chans, reqs)
	case "control":
		username := conn.Permissions.Extensions["ruka-username"]
		go ssh.DiscardRequests(reqs)
		s.handleControlSession(username, chans)
	case "jump":
		handleJumpSession(s.s, conn, chans, reqs)
	default:
		return fmt.Errorf("unsupported session type: %s", st)
	}

	return nil
}

func (s *Server) authenticatePubKey() func(ssh.ConnMetadata, ssh.PublicKey) (*ssh.Permissions, error) {
	return func(cm ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
		switch cm.User() {
		case "node":
			nodename, err := s.s.FetchNodenameByKey(key.Marshal())
			if err != nil {
				return nil, errors.Wrap(err, "authentication failed")
			}
			return &ssh.Permissions{
				Extensions: map[string]string{
					"ruka-session-type": cm.User(),
					"ruka-nodename":     nodename,
				},
			}, nil
		case "control", "jump":
			if string(key.Marshal()) == string(s.rootKey.Marshal()) {
				return &ssh.Permissions{Extensions: map[string]string{
					"ruka-session-type": cm.User(),
					"ruka-username":     "root",
				}}, nil
			}

			username, err := s.s.FetchUsernameByKey(key.Marshal())
			if err != nil {
				return nil, errors.Wrap(err, "authentication failed")
			}

			return &ssh.Permissions{Extensions: map[string]string{
				"ruka-session-type": cm.User(),
				"ruka-username":     username,
			}}, nil
		default:
			return nil, fmt.Errorf("authentication failed: invalid user: %s", cm.User())
		}
	}
}
