package main

import (
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli"

	m "github.com/armedev/simple-go-service"
)

func main() {
	app := &cli.App{
		Name:  "albums",
		Usage: "albums cli",
		Commands: []cli.Command{
			{
				Name:    "get",
				Aliases: []string{"g"},
				Usage:   "get a album",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:     "id",
						Usage:    "id of the album",
						Required: true,
					},
				},
				Action: func(cCtx *cli.Context) error {
					id := cCtx.String("id")

					db := m.CustomDb{
						path: "./data/albums",
					}

					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
