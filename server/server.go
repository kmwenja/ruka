package server

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"

	"github.com/kmwenja/ruka/server/control"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
)

// Config TODO
type Config struct {
	Addr        string
	HostKeyFile string
	RootKeyFile string
}

// Backend TODO
type Backend interface {
	control.ShellBackend
	Authenticate(st SessionType, key []byte) (string, error)
}

// Start is the entry point for the server
func Start(backend Backend, cfg *Config) error {
	hb, err := ioutil.ReadFile(cfg.HostKeyFile)
	if err != nil {
		return errors.Wrapf(err, "could not open host private key: %s", cfg.HostKeyFile)
	}
	hk, err := ssh.ParsePrivateKey(hb)
	if err != nil {
		return errors.Wrapf(err, "could not parse host private key: %s", cfg.HostKeyFile)
	}

	rb, err := ioutil.ReadFile(cfg.RootKeyFile)
	if err != nil {
		return errors.Wrapf(err, "could not open root public key: %s", cfg.RootKeyFile)
	}
	rk, _, _, _, err := ssh.ParseAuthorizedKey(rb)
	if err != nil {
		return errors.Wrapf(err, "could not parse root public key: %s", cfg.RootKeyFile)
	}

	ssc := &ssh.ServerConfig{
		MaxAuthTries:      3,
		PublicKeyCallback: authenticatePubKey(backend, rk),
		AuthLogCallback:   nil,                // TODO log auths
		ServerVersion:     "SSH-2.0-ruka-0.1", // TODO use actual version
		BannerCallback:    nil,                // TODO nice banner
	}
	ssc.AddHostKey(hk)

	l, err := net.Listen("tcp", cfg.Addr)
	if err != nil {
		return errors.Wrapf(err, "cannot listen to %s", cfg.Addr)
	}
	defer l.Close()
	log.Printf("Listening to %s", cfg.Addr)

	for {
		c, err := l.Accept()
		if err != nil {
			log.Printf("Error: could not accept connection: %v", err)
			continue
		}

		go func(c net.Conn) {
			conn, chans, reqs, err := ssh.NewServerConn(c, ssc)
			if err != nil {
				log.Printf("Error: could not initialize ssh connection: %v", err)
				return
			}
			defer conn.Close()

			t, err := SessionTypeFromString(conn.Permissions.Extensions["ruka-session-type"])
			if err != nil {
				log.Printf("Error: could not get session type: %v", err)
				return
			}

			switch t {
			case SessionType_NODE:
				handleNodeSession(backend, conn, chans, reqs)
			case SessionType_CONTROL:
				handleControlSession(backend, conn, chans, reqs)
			case SessionType_JUMP:
				handleJumpSession(backend, conn, chans, reqs)
			}
		}(c)
	}
}

func authenticatePubKey(backend Backend, rootKey ssh.PublicKey) func(ssh.ConnMetadata, ssh.PublicKey) (*ssh.Permissions, error) {
	return func(cm ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
		t, err := SessionTypeFromString(cm.User())
		if err != nil {
			return nil, errors.Wrap(err, "Authentication failed")
		}

		if string(key.Marshal()) == string(rootKey.Marshal()) {
			if t != SessionType_CONTROL {
				newErr := fmt.Errorf("Root only allowed in control sessions")
				return nil, errors.Wrap(newErr, "Authentication failed")
			}

			return &ssh.Permissions{
				Extensions: map[string]string{
					"ruka-session-type": t.String(),
					"ruka-username":     "root",
				},
			}, nil
		}

		username, err := backend.Authenticate(t, key.Marshal())
		if err != nil {
			return nil, errors.Wrap(err, "Authentication failed")
		}

		return &ssh.Permissions{
			Extensions: map[string]string{
				"ruka-session-type": t.String(),
				"ruka-username":     username,
			},
		}, nil
	}
}
