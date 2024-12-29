// main.go
package main

import (
	"log"
	"os"

	cli "github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "quic-ssh",
		Usage: "A QUIC-based SSH client and server",
		Commands: []*cli.Command{
			{
				Name:    "server",
				Aliases: []string{"s"},
				Usage:   "Start the QUIC-based SSH server",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "bind",
						Usage:    "bind address, eg. 0.0.0.0:4242",
						Required: true,
						Value:    "0.0.0.0:4242",
					},
					&cli.StringFlag{
						Name:     "cert",
						Usage:    "TLS cert path",
						Required: false,
						Value:    "server.crt",
					},
					&cli.StringFlag{
						Name:     "key",
						Usage:    "TLS key pass",
						Required: false,
						Value:    "server.key",
					},
				},
				Action: server,
			},
			{
				Name:    "client",
				Aliases: []string{"c"},
				Usage:   "Start the QUIC-based SSH client",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "addr",
						Usage:    "server ip address, eg. localhost:4242",
						Required: true,
						Value:    "localhost:4242",
					},
				},
				Action: client,
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatalf("run failure: %v", err)
	}
}
