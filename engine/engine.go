package engine

import (
	"fmt"
	"os"

	"github.com/denkhaus/tcgl/applog"
)

////////////////////////////////////////////////////////////////////////////////
type Engine struct {
}

////////////////////////////////////////////////////////////////////////////////
type CtxBase struct {
	Username string
	Password string
	Host     string
	Port     int
}

type EngineFunc func(engine *Engine) error

////////////////////////////////////////////////////////////////////////////////
func (e *Engine) Execute(fn EngineFunc) error {
	return fn(e)
}

////////////////////////////////////////////////////////////////////////////////
func (e *Engine) initDataStore(reset bool, args ...interface{}) (*DataStore, error) {
	dbPath, err := getDBPathByArgs(args...)
	if err != nil {
		return nil, fmt.Errorf("Store::GetDBPathByArgs::%s", err)
	}

	if reset {
		applog.Infof("Store::Reset database %s", dbPath)
		os.Remove(dbPath)
	}

	store, err := NewDatastore(dbPath, "UIDMap")
	if err != nil {
		return nil, fmt.Errorf("Store::%s", err)
	}

	if reset {
		applog.Infof("Store::Start with new database %s", dbPath)
	} else {
		applog.Infof("Store::Use database %s", dbPath)
	}

	return store, nil
}

////////////////////////////////////////////////////////////////////////////////
func NewEngine() (*Engine, error) {
	return new(Engine), nil
}
