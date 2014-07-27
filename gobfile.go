package gobfile

import (
	"encoding/gob"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"os"
	"reflect"
	"strconv"
	"sync"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type File struct {
	Obj      interface{}
	portLock net.Listener
	cbs      chan func()
	path     string
}

func New(obj interface{}, path string, lockPort int) (*File, error) {
	// check object
	if reflect.TypeOf(obj).Kind() != reflect.Ptr {
		return nil, errors.New("object must be a pointer")
	}

	// init
	file := &File{
		Obj:  obj,
		cbs:  make(chan func()),
		path: path,
	}

	// try lock
	ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", lockPort))
	if err != nil {
		return nil, err
	}
	file.portLock = ln

	// try load from file
	dbFile, err := os.Open(path)
	if err == nil {
		defer dbFile.Close()
		err = gob.NewDecoder(dbFile).Decode(file.Obj)
		if err != nil {
			return nil, err
		}
	}

	// loop
	go func() {
		for {
			cb, ok := <-file.cbs
			if !ok {
				return
			}
			cb()
		}
	}()

	return file, nil
}

func (f *File) Save() (err error) {
	var done sync.Mutex
	done.Lock()
	f.cbs <- func() {
		defer done.Unlock()
		tmpPath := f.path + "." + strconv.FormatInt(rand.Int63(), 10)
		var tmpF *os.File
		tmpF, err = os.Create(tmpPath)
		if err != nil {
			return
		}
		defer tmpF.Close()
		err = gob.NewEncoder(tmpF).Encode(f.Obj)
		if err != nil {
			return
		}
		err = os.Rename(tmpPath, f.path)
		if err != nil {
			return
		}
	}
	done.Lock()
	return
}

func (f *File) Close() {
	close(f.cbs)
	f.portLock.Close()
}
