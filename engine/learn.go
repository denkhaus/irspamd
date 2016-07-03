package engine

import (
	"bytes"

	"github.com/denkhaus/irspamd/rspamd"
	"github.com/denkhaus/tcgl/applog"
	"github.com/juju/errors"
	"github.com/tgulacsi/imapclient"
)

////////////////////////////////////////////////////////////////////////////////
type LearnCtx struct {
	CtxBase
	LearnBox string
	FnString string
	MarkSeen bool
}

////////////////////////////////////////////////////////////////////////////////
func (ctx LearnCtx) Print() {
	applog.Infof("//////////////////////////// Learn //////////////////////////////////")
	applog.Infof("// Use Connection %s:%d/user=%s", ctx.Host, ctx.Port, ctx.Username)

	if ctx.FnString == "learn_ham" {
		applog.Infof("// Learn ham from %s", ctx.LearnBox)
	} else if ctx.FnString == "learn_spam" {
		applog.Infof("// Learn spam from %s", ctx.LearnBox)
	}
	if ctx.MarkSeen {
		applog.Infof("// Mark lerned Emails as seen")
	}

	applog.Infof("////////////////////////////////////////////////////////////////////")
}

////////////////////////////////////////////////////////////////////////////////
func (e *Engine) Learn(ctx LearnCtx) error {
	ctx.Print()

	c := imapclient.NewClient(ctx.Host, ctx.Port, ctx.Username, ctx.Password)
	if err := c.Connect(); err != nil {
		return errors.Annotate(err, "connect")
	}
	defer c.Close(false)

	uids, err := c.List(ctx.LearnBox, "", false)
	if err != nil {
		return errors.Annotate(err, "list new")
	}

	applog.Infof("Learn %d new messages.", len(uids))

	var body bytes.Buffer
	for _, uid := range uids {
		body.Reset()
		if _, err := c.ReadTo(&body, uid); err != nil {
			return errors.Annotate(err, "read body")
		}

		resp, err := rspamd.Learn(ctx.FnString, &body)
		if err != nil {
			return errors.Annotate(err, "learn")
		}

		resp.Report(uid)

		if ctx.MarkSeen {
			if err := c.Mark(uid, true); err != nil {
				return errors.Annotate(err, "mark seen")
			}
		}
	}

	return nil
}
