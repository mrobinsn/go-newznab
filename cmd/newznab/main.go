package main

import (
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
)

var (
	app = initApp()
)

func initApp() *cli.App {
	app := cli.NewApp()

	app.Name = "newznab"
	app.Version = "0.1.0"
	app.Usage = ""
	app.Before = globalSetup
	app.Commands = []cli.Command{}
	app.Authors = []cli.Author{
		{Name: "Michael Robinson", Email: "mrobinson@outlook.com"},
	}

	app.Flags = []cli.Flag{}

	return app
}

func main() {
	if err := app.Run(os.Args); err != nil {
		log.WithError(err).Fatal("app returned error")
	}
}

func globalSetup(c *cli.Context) error {
	return nil
}
