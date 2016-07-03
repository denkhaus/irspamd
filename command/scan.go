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
			cli.StringFlag{
				Name: "inbox, i", Value: "INBOX",
				Usage: "Name of the box to be scanned.",
			},
			cli.StringFlag{
				Name: "spambox, s", Value: "Spam",
				Usage: "Name of the box to move spam messages to.",
			},
			cli.StringFlag{
				Name: "hambox, m", Value: "",
				Usage: "Name of the box to move ham messages to. If no hambox is given, ham remains in inbox.",
			},
			cli.BoolFlag{
				Name:  "expunge, e",
				Usage: "Expunge all spam messages from inbox after scan has finished.",
			},
			cli.BoolFlag{
				Name:  "force, f",
				Usage: "Forces processing if message is already checked.",
			},
		},
		Action: func(ctx *cli.Context) error {
			return c.Execute(func(eng *engine.Engine) error {

				sCtx := engine.CheckCtx{
					SpamBox: ctx.String("spambox"),
					HamBox:  ctx.String("hambox"),
					InBox:   ctx.String("inbox"),
					Expunge: ctx.Bool("expunge"),
					Force:   ctx.Bool("force"),
					CtxBase: engine.CtxBase{
						Host:     ctx.GlobalString("host"),
						Port:     ctx.GlobalInt("port"),
						Username: ctx.GlobalString("user"),
						Password: ctx.GlobalString("pass"),
					},
				}

				return eng.Check(sCtx)
			}, ctx)
		},
	})
}
