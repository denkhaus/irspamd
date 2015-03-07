package engine

import (
	"bytes"
	"fmt"
	"time"

	"github.com/denkhaus/imapclient"
	"github.com/denkhaus/irspamd/rspamd"
	"github.com/denkhaus/tcgl/applog"
)

type SpamRec struct {
	Uid         uint32
	LastScanned time.Time
	Error       string
	Response    rspamd.Response
}

////////////////////////////////////////////////////////////////////////////////
func (sr SpamRec) IsSpam() bool {
	return sr.Response.Spam
}

////////////////////////////////////////////////////////////////////////////////
func (sr SpamRec) Score() float64 {
	return sr.Response.Score
}

////////////////////////////////////////////////////////////////////////////////
func (sr SpamRec) FmtSpamKey(pattern string) string {
	x := int(sr.Response.Score/sr.Response.Threshold*100) / 10 * 10
	return fmt.Sprintf(`%s%d`, pattern, x)
}

////////////////////////////////////////////////////////////////////////////////
type ScanContext struct {
	ContextBase

	Expunge bool
	SpamBox string
	HamBox  string
	InBox   string
}

////////////////////////////////////////////////////////////////////////////////
func (ctx ScanContext) Print() {
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
func (e *Engine) Scan(ctx ScanContext) error {
	ctx.Print()

	c := imapclient.NewClient(ctx.Host, ctx.Port, ctx.Username, ctx.Password)
	if err := c.Connect(); err != nil {
		return fmt.Errorf("Imap::Connect::%s", err)
	}

	defer c.Close(ctx.Expunge)

	store, err := e.initDataStore(ctx.Ephemeral, ctx.Host, ctx.Port, ctx.Username,
		ctx.InBox, ctx.HamBox, ctx.SpamBox)
	if err != nil {
		return err
	}
	defer store.Close()

	uids, err := c.ListNew(ctx.InBox, "")
	if err != nil {
		return fmt.Errorf("Imap::ListNew::%s", err)
	}

	msgCount := len(uids)

	if msgCount > 0 {
		applog.Infof("Scan::Check %d new messages for spam.", msgCount)
	} else {
		applog.Infof("Scan::no new messages to scan.")
	}

	var body bytes.Buffer

	for _, uid := range uids {
		rec := SpamRec{}
		if err := store.GetRecordById(uid, &rec); err != nil {
			applog.Errorf("Store::GetSpamRecordById::%s", err)
		} else if rec.Uid == uid {
			applog.Infof("Scan::Message %d already processed. Is spam?: %t, score: %.2f",
				uid, rec.IsSpam(), rec.Score())
			continue
		}

		body.Reset()
		if _, err := c.ReadTo(&body, uid); err != nil {
			return fmt.Errorf("Imap::ReadBody::%s", err)
		}

		rec = SpamRec{Uid: uid, LastScanned: time.Now().UTC()}
		if resp, err := rspamd.CheckSpam(e.RspamdConfig, body.Bytes()); err != nil {
			applog.Errorf("Rspamd::CheckSpam::uid %d, %s", uid, err)
			rec.Error = err.Error()
		} else {

			// remove eventually set RSPAMD_SCORE flag
			if err := c.SetFlagRegex(uid, "RSPAMD_SCORE_.*", false); err != nil {
				return fmt.Errorf("Imap::ClearFlag::%s", err)
			}

			rec.Response = *resp
			key := rec.FmtSpamKey("RSPAMD_SCORE_")
			if err := c.SetFlag(uid, key, true); err != nil {
				return fmt.Errorf("Imap::SetKeyword::%s", err)
			}

			if rec.IsSpam() {
				applog.Infof("Scan::Mail %d is spam", uid)
				if err := c.Move(uid, ctx.SpamBox); err != nil {
					return fmt.Errorf("Imap::MoveToSpamFolder::%s", err)
				}
			} else {
				applog.Infof("Scan::Mail %d is clean", uid)
				if ctx.HamBox != "" {
					if err := c.Move(uid, ctx.HamBox); err != nil {
						return fmt.Errorf("Imap::MoveToHamFolder::%s", err)
					}
				}
			}
		}

		if err := store.PutRecord(uid, rec); err != nil {
			return fmt.Errorf("Store::PutRecord::%s", err)
		}
	}

	return nil
}
