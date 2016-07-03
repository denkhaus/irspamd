package command

import (
	"github.com/codegangsta/cli"
	"github.com/denkhaus/irspamd/engine"
	"github.com/juju/errors"
)

func (c *Commander) NewLearnCommand() {
	c.Register(cli.Command{
		Name:  "learn",
		Usage: "Learn ham or spam from given IMAP box.",
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "markseen, s",
				Usage: "Mark Emails as seen",
			},
		},
		Subcommands: []cli.Command{
			{
				Name:  "ham",
				Usage: "Learn ham from learnbox.",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:   "learnbox, l",
						Value:  "",
						Usage:  "Name of the box to be scanned for learning. Required",
						EnvVar: "",
					},
				},
				Action: func(ctx *cli.Context) error {
					return c.Execute(func(eng *engine.Engine) error {
						return buildContextAndLearn(ctx, eng, "learn_ham")
					}, ctx)
				},
			},
			{
				Name:  "spam",
				Usage: "Learn spam from learnbox.",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:   "learnbox, l",
						Value:  "",
						Usage:  "Name of the box to be scanned for learning. Required",
						EnvVar: "",
					},
				},
				Action: func(ctx *cli.Context) error {
					return c.Execute(func(eng *engine.Engine) error {
						return buildContextAndLearn(ctx, eng, "learn_spam")
					}, ctx)
				},
			},
		},
	})
}

////////////////////////////////////////////////////////////////////////////////
func buildContextAndLearn(ctx *cli.Context, e *engine.Engine, fnString string) error {
	if !ctx.IsSet("learnbox") {
		return errors.New("Learn::learnbox is not set. Value is mandatory.")
	}

	lCtx := engine.LearnCtx{
		LearnBox: ctx.String("learnbox"),
		FnString: fnString,
		MarkSeen: ctx.GlobalBool("markseen"),
		CtxBase: engine.CtxBase{
			Host:     ctx.GlobalString("host"),
			Port:     ctx.GlobalInt("port"),
			Username: ctx.GlobalString("user"),
			Password: ctx.GlobalString("pass"),
		},
	}

	return e.Learn(lCtx)
}
