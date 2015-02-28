package engine

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"os"
	"os/signal"
	"os/user"
	"path"
	"reflect"
	"strconv"
	"syscall"

	"github.com/boltdb/bolt"
	"github.com/denkhaus/tcgl/applog"
)

////////////////////////////////////////////////////////////////////////////////
type DataStore struct {
	db         *bolt.DB
	bucketName []byte
}

////////////////////////////////////////////////////////////////////////////////
var (
	ErrNotFound = fmt.Errorf("Datastore:: value not found")
)

////////////////////////////////////////////////////////////////////////////////
func NewDatastore(path, bucket string) (*DataStore, error) {
	db, err := bolt.Open(path, 0600, nil)
	if err != nil {
		return nil, err
	}

	if err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucket))
		return err
	}); err != nil {
		return nil, err
	}

	store := &DataStore{
		db:         db,
		bucketName: []byte(bucket),
	}

	go store.signalHandler()
	return store, nil
}

////////////////////////////////////////////////////////////////////////////////
func (d *DataStore) signalHandler() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	<-sigChan

	if d.db != nil {
		d.db.Close()
		d.db = nil
		applog.Infof("Store::DB successfull closed")
	}
	os.Exit(0)
}

////////////////////////////////////////////////////////////////////////////////
func (d *DataStore) Close() (err error) {
	err = d.db.Close()
	d.db = nil
	return
}

////////////////////////////////////////////////////////////////////////////////
func (d *DataStore) Delete(key []byte) error {
	return d.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(d.bucketName).Delete(key)
	})
}

////////////////////////////////////////////////////////////////////////////////
func (d *DataStore) Get(key []byte) (interface{}, error) {
	var out []byte
	err := d.db.View(func(tx *bolt.Tx) error {
		mmval := tx.Bucket(d.bucketName).Get(key)
		if mmval == nil {
			return ErrNotFound
		}

		out = make([]byte, len(mmval))
		copy(out, mmval)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return out, err
}

////////////////////////////////////////////////////////////////////////////////
func (d *DataStore) Has(key []byte) (bool, error) {
	var found bool
	err := d.db.View(func(tx *bolt.Tx) error {
		val := tx.Bucket(d.bucketName).Get(key)
		found = (val != nil)
		return nil
	})
	return found, err
}

////////////////////////////////////////////////////////////////////////////////
func (d *DataStore) PutRecord(uid uint32, rec interface{}) error {
	if err := d.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(d.bucketName))

		var buf bytes.Buffer
		enc := gob.NewEncoder(&buf)
		if err := enc.Encode(rec); err != nil {
			return err
		}

		key := []byte(strconv.Itoa(int(uid)))
		return b.Put(key, buf.Bytes())
	}); err != nil {
		return err
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////
func (d *DataStore) GetRecordById(uid uint32, rec interface{}) error {
	sv := reflect.ValueOf(rec)
	if sv.Kind() != reflect.Ptr {
		return fmt.Errorf("GetRecordById:: rec must be a value ptr.")
	}
	if sv.IsNil() {
		return fmt.Errorf("GetRecordById:: rec must not be nil.")
	}

	var buf bytes.Buffer
	dec := gob.NewDecoder(&buf)
	err := d.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(d.bucketName))
		k := []byte(strconv.Itoa(int(uid)))
		if v := b.Get(k); v != nil {
			if _, err := buf.Write(v); err != nil {
				return err
			}
			if err := dec.Decode(rec); err != nil {
				return err
			}
		}
		return nil
	})

	return err
}

////////////////////////////////////////////////////////////////////////////////
func GetDBPathByArgs(args ...interface{}) (string, error) {
	hasher := md5.New()
	for _, arg := range args {
		switch v := arg.(type) {
		case string:
			hasher.Write([]byte(v))
		case int:
			buf := new(bytes.Buffer)
			if err := binary.Write(buf, binary.LittleEndian, int32(v)); err != nil {
				return "", err
			}
			hasher.Write(buf.Bytes())
		default:
			fmt.Errorf("GetDBPathByArgs::Type not supported")
		}
	}

	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	return path.Join(usr.HomeDir, fmt.Sprintf("imapspam%s.db",
		hex.EncodeToString(hasher.Sum(nil)))), nil
}
