package main

import (
	"os"

	"github.com/codegangsta/cli"
	"github.com/davecgh/go-spew/spew"
	"github.com/denkhaus/irspamd/command"
	"github.com/denkhaus/tcgl/applog"
)

func Inspect(args ...interface{}) {
	spew.Dump(args)
}

func main() {

	app := cli.NewApp()
	app.Name = "irspamd"
	app.Version = AppVersion
	app.Usage = "A command line app that scans your IMAP mail for spam."
	app.Flags = []cli.Flag{
		cli.StringFlag{"host, H", "localhost", "Host to connect to.", ""},
		cli.IntFlag{"port, P", 993, "Port number to connect to.", ""},
		cli.StringFlag{"user, u", "", "Your username at host", ""},
		cli.StringFlag{"pass, p", "", "Your IMAP password. For security reasons prefer IMAP_PASSWORD='yourpassword'", "IMAP_PASSWORD"},
		cli.BoolFlag{"reset, r", "Clear database before run", ""},
		//cli.StringSliceFlag{"peers, C", &cli.StringSlice{}, "a comma-delimited list of machine addresses in the cluster (default: {\"127.0.0.1:4001\"})"},
	}

	if cmdr, err := command.NewCommander(app); err != nil {
		applog.Errorf("Startup error:: %s", err.Error())
		return
	} else {
		cmdr.Run(os.Args)
	}
}
