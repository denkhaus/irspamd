package engine

import (
	"bytes"
	"fmt"
	"time"

	"github.com/denkhaus/imapclient"
	"github.com/denkhaus/irspamd/rspamd"
	"github.com/denkhaus/tcgl/applog"
)

type LearnRec struct {
	Uid         uint32
	LastLearned time.Time
	Error       string
	Response    string
}

////////////////////////////////////////////////////////////////////////////////
type LearnContext struct {
	ContextBase
	LearnBox string
	FnString string
}

////////////////////////////////////////////////////////////////////////////////
func (ctx LearnContext) Print() {
	applog.Infof("//////////////////////////// Learn //////////////////////////////////")
	applog.Infof("// Use Connection %s:%d/user=%s", ctx.Host, ctx.Port, ctx.Username)

	if ctx.FnString == "learn_ham" {
		applog.Infof("// Learn ham from %s", ctx.LearnBox)
	} else if ctx.FnString == "learn_spam" {
		applog.Infof("// Learn spam from %s", ctx.LearnBox)
	}

	applog.Infof("////////////////////////////////////////////////////////////////////")
}

////////////////////////////////////////////////////////////////////////////////
func (e *Engine) Learn(ctx LearnContext) error {
	ctx.Print()

	c := imapclient.NewClient(ctx.Host, ctx.Port, ctx.Username, ctx.Password)
	if err := c.Connect(); err != nil {
		return fmt.Errorf("Imap::Connect::%s", err)
	}
	defer c.Close(false)

	store, err := e.initDataStore(ctx.ResetDb, ctx.Host,
		ctx.Port, ctx.Username, ctx.LearnBox)
	if err != nil {
		return err
	}
	defer store.Close()

	uids, err := c.ListNew(ctx.LearnBox, "")
	if err != nil {
		return fmt.Errorf("Imap::ListNew::%s", err)
	}

	applog.Infof("Learn::Learn %d new messages.", len(uids))

	var body bytes.Buffer
	for _, uid := range uids {
		rec := LearnRec{}
		if err := store.GetRecordById(uid, &rec); err != nil {
			applog.Errorf("Store::GetRecordById::%s", err)
		} else if rec.Uid == uid {
			applog.Infof("Learn::Message %d already learned at %s", uid, rec.LastLearned)
			continue
		}

		body.Reset()
		if _, err := c.ReadTo(&body, uid); err != nil {
			return fmt.Errorf("Imap::ReadBody::%s", err)
		}

		rec = LearnRec{Uid: uid, LastLearned: time.Now().UTC()}
		if resp, err := rspamd.Learn(ctx.FnString, &body); err != nil {
			applog.Errorf("Rspamd::Learn::uid %d, %s", uid, err)
			rec.Error = err.Error()
		} else {
			rec.Response = resp
		}
		if err := store.PutRecord(uid, rec); err != nil {
			return fmt.Errorf("Store::PutRecord::%s", err)
		}
	}

	return nil
}
