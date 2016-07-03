package engine

import (
	"bytes"
	"strings"

	"github.com/denkhaus/imapclient"
	"github.com/denkhaus/irspamd/rspamd"
	"github.com/denkhaus/tcgl/applog"
	"github.com/juju/errors"
)

////////////////////////////////////////////////////////////////////////////////
type CheckCtx struct {
	CtxBase
	Expunge bool
	Force   bool
	SpamBox string
	HamBox  string
	InBox   string
}

////////////////////////////////////////////////////////////////////////////////
func (ctx CheckCtx) Print() {
	applog.Infof("//////////////////////////// Scan //////////////////////////////////")
	applog.Infof("// Use Connection %s:%d/user=%s", ctx.Host, ctx.Port, ctx.Username)
	applog.Infof("// Inbox: %s, Spambox: %s", ctx.InBox, ctx.SpamBox)

	if ctx.HamBox != "" {
		applog.Infof("// Ham is moved to %s", ctx.HamBox)
	} else {
		applog.Infof("// Ham remains in %s", ctx.InBox)
	}

	if ctx.Expunge {
		applog.Infof("// Expunge is on.")
	} else {
		applog.Infof("// Expunge is off.")
	}

	applog.Infof("////////////////////////////////////////////////////////////////////")
}

////////////////////////////////////////////////////////////////////////////////
func (e *Engine) Check(ctx CheckCtx) error {
	ctx.Print()

	c := imapclient.NewClient(ctx.Host, ctx.Port, ctx.Username, ctx.Password)
	if err := c.Connect(); err != nil {
		return errors.Annotate(err, "connect")
	}

	defer c.Close(ctx.Expunge)

	uids, err := c.List(ctx.InBox, "", false)
	if err != nil {
		return errors.Annotate(err, "list new")
	}

	msgCount := len(uids)

	if msgCount > 0 {
		applog.Infof("Scan::Check %d new messages for spam.", msgCount)
	} else {
		applog.Infof("Scan::no new messages to scan.")
	}

	var body bytes.Buffer

	for _, uid := range uids {
		flagSet, err := c.GetFlags(uid)
		if err != nil {
			return errors.Annotate(err, "get flags")
		}

		scored := false
		for k, _ := range flagSet {
			if strings.HasPrefix(k, "RSPAMD_SCORE_") {
				scored = true
				break
			}
		}

		if !ctx.Force && scored {
			continue
		}

		body.Reset()
		if _, err := c.ReadTo(&body, uid); err != nil {
			return errors.Annotate(err, "read mail body")
		}

		resp, err := rspamd.Check(&body)
		if err != nil {
			return errors.Annotatef(err, "check uid %d", uid)
		}

		resp.Report(uid)
		// remove eventually set RSPAMD_SCORE flag
		if err := c.SetFlagRegex(uid, "RSPAMD_SCORE_.*", false); err != nil {
			return errors.Annotate(err, "clear flag")
		}

		key := resp.FmtScore("RSPAMD_SCORE_")
		if err := c.SetFlag(uid, key, true); err != nil {
			return errors.Annotate(err, "set flag")
		}

		if resp.Spam {
			if ctx.SpamBox != "" {
				if err := c.Move(uid, ctx.SpamBox); err != nil {
					return errors.Annotatef(err, "move uid %d to spam folder", uid)
				}
			}
		} else {
			if ctx.HamBox != "" {
				if err := c.Move(uid, ctx.HamBox); err != nil {
					return errors.Annotatef(err, "move uid %d to ham folder", uid)
				}
			}
		}
	}

	return nil
}
