package command

import (
	"github.com/codegangsta/cli"
	"github.com/denkhaus/irspamd/engine"
	"github.com/denkhaus/irspamd/rspamd"
	"github.com/denkhaus/tcgl/applog"
)

type Commander struct {
	engine *engine.Engine
	app    *cli.App
}

///////////////////////////////////////////////////////////////////////////////////////////////
func (c *Commander) Execute(fn engine.EngineFunc, ctx *cli.Context) {
	if err := c.engine.Execute(fn); err != nil {
		applog.Errorf("Execution error::%s", err.Error())
	}
}

///////////////////////////////////////////////////////////////////////////////////////////////
func NewCommander(app *cli.App) (*Commander, error) {
	cmd := &Commander{app: app}

	saConfig := &rspamd.Config{
		Ip:      "127.0.0.1",
		Port:    11333,
		Timeout: 10,
	}

	if engine, err := engine.NewEngine(saConfig); err != nil {
		return nil, err
	} else {
		cmd.engine = engine
	}

	cmd.NewScanCommand()
	cmd.NewLearnCommand()
	return cmd, nil
}

///////////////////////////////////////////////////////////////////////////////////////////////
func (c *Commander) Register(cmd cli.Command) {
	c.app.Commands = append(c.app.Commands, cmd)
}

///////////////////////////////////////////////////////////////////////////////////////////////
func (c *Commander) Run(args []string) error {
	return c.app.Run(args)
}
