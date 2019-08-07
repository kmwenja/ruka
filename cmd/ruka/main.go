package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/kmwenja/ruka/node"
	"github.com/kmwenja/ruka/server"
	"github.com/kmwenja/ruka/server/stores/storm"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "ruka"

	app.Commands = []cli.Command{
		serverCmd(),
		nodeCmd(),
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Printf("Error while parsing arguments: %v\n", err)
		os.Exit(1)
	}
}

func serverCmd() cli.Command {
	c := cli.Command{
		Name:  "server",
		Usage: "manage a ruka server",
		Subcommands: []cli.Command{
			{
				Name:  "init",
				Usage: "setup a working bare minimum environment for a ruka server",
				Action: func(c *cli.Context) error {
					fmt.Println("Setup ruka server!")
					return nil
				},
			},
			{
				Name:  "start",
				Usage: "start ruka server based off the current working directory",
				Action: func(c *cli.Context) error {
					store, err := storm.New("/tmp/data")
					if err != nil {
						return errors.Wrapf(err, "could not init backend")
					}

					s, err := server.New(server.Config{
						Store:       store,
						HostKeyFile: "/tmp/test",
						RootKeyFile: "/tmp/test.pub",
					})
					if err != nil {
						return errors.Wrapf(err, "could not initialize server")
					}

					addr := ":2022"
					l, err := net.Listen("tcp", addr)
					if err != nil {
						return errors.Wrapf(err, "cannot listen to %s", addr)
					}
					defer l.Close()
					log.Printf("Listening to %s", addr)
					return s.Serve(l)
				},
			},
		},
	}
	return c
}

func nodeCmd() cli.Command {
	c := cli.Command{
		Name:  "node",
		Usage: "manage a ruka node",
		Subcommands: []cli.Command{
			{
				Name:  "init",
				Usage: "setup a working bare minimum environment for a ruka node",
				Action: func(c *cli.Context) error {
					fmt.Println("Setup ruka node!")
					return nil
				},
			},
			{
				Name:  "start",
				Usage: "start ruka node based off the current working directory",
				Action: func(c *cli.Context) error {
					n, err := node.New(node.Config{
						NodeKeyFile: "/tmp/test.pub",
						HostKeyFile: "/tmp/test",
					})
					if err != nil {
						return errors.Wrapf(err, "could not initiate node")
					}

					addr := "127.0.0.1:2022"
					conn, err := net.Dial("tcp", addr)
					if err != nil {
						return errors.Wrapf(err, "could not connect to server at `%s`", addr)
					}
					defer conn.Close()
					return n.Serve(addr, conn)
				},
			},
		},
	}
	return c
}
