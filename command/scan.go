package command

import (
	"github.com/codegangsta/cli"
	"github.com/denkhaus/irspamd/engine"
)

func (c *Commander) NewScanCommand() {
	c.Register(cli.Command{
		Name:  "scan",
		Usage: "Scans the given inbox and moves spam messages to specified spambox.",
		Flags: []cli.Flag{
			cli.StringFlag{"inbox, i", "INBOX", "Name of the box to be scanned.", ""},
			cli.StringFlag{"spambox, s", "Spam", "Name of the box to move spam messages to.", ""},
			cli.StringFlag{"hambox, m", "", "Name of the box to move ham messages to. If no hambox is given, ham remains in inbox.", ""},
			cli.BoolFlag{"expunge, e", "Expunge all spam messages from inbox after scan has finished.", ""},
		},
		Action: func(ctx *cli.Context) {
			c.Execute(func(eng *engine.Engine) error {

				sCtx := engine.ScanContext{
					SpamBox: ctx.String("spambox"),
					HamBox:  ctx.String("hambox"),
					InBox:   ctx.String("inbox"),
					Expunge: ctx.Bool("expunge"),
					ContextBase: engine.ContextBase{
						Host:     ctx.GlobalString("host"),
						Port:     ctx.GlobalInt("port"),
						Username: ctx.GlobalString("user"),
						Password: ctx.GlobalString("pass"),
						ResetDb:  ctx.GlobalBool("reset"),
					},
				}

				return eng.Scan(sCtx)
			}, ctx)
		},
	})
}
