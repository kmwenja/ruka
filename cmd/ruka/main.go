package main

import (
	"fmt"
	"os"

	"github.com/kmwenja/ruka/server"
	"github.com/kmwenja/ruka/server/backends/storm"
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
					scfg := &server.Config{
						Addr:        ":2022",
						HostKeyFile: "/tmp/test",
						RootKeyFile: "/tmp/test.pub",
					}
					backend, err := storm.New("/tmp/data")
					if err != nil {
						return errors.Wrapf(err, "could not init backend")
					}
					return server.Start(backend, scfg)
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
					fmt.Println("Start ruka node!")
					return nil
				},
			},
		},
	}
	return c
}
