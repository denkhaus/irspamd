package command

import (
	"fmt"

	"github.com/codegangsta/cli"
	"github.com/denkhaus/irspamd/engine"
)

func (c *Commander) NewLearnCommand() {
	c.Register(cli.Command{
		Name:  "learn",
		Usage: "Learn ham or spam from given IMAP box.",
		Subcommands: []cli.Command{
			{
				Name:  "ham",
				Usage: "Learn ham from learnbox.",
				Flags: []cli.Flag{
					cli.StringFlag{"learnbox, l", "", "Name of the box to be scanned for learning. Required", ""},
				},
				Action: func(ctx *cli.Context) {
					c.Execute(func(eng *engine.Engine) error {
						return buildContextAndLearn(ctx, eng, "learn_ham")
					}, ctx)
				},
			},
			{
				Name:  "spam",
				Usage: "Learn spam from learnbox.",
				Flags: []cli.Flag{
					cli.StringFlag{"learnbox, l", "", "Name of the box to be scanned for learning. Required", ""},
				},
				Action: func(ctx *cli.Context) {
					c.Execute(func(eng *engine.Engine) error {
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
		return fmt.Errorf("Learn::learnbox is not set. Value is mandatory.")
	}

	lCtx := engine.LearnContext{
		LearnBox: ctx.String("learnbox"),
		FnString: fnString,
		ContextBase: engine.ContextBase{
			Host:      ctx.GlobalString("host"),
			Port:      ctx.GlobalInt("port"),
			Username:  ctx.GlobalString("user"),
			Password:  ctx.GlobalString("pass"),
			Ephemeral: ctx.GlobalBool("ephemeral"),
		},
	}

	return e.Learn(lCtx)
}
