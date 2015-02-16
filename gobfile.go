package gobfile

import (
	crand "crypto/rand"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"math/rand"
	"os"
	"reflect"
	"strconv"
	"sync"
	"time"
)

func init() {
	var seed int64
	binary.Read(crand.Reader, binary.LittleEndian, &seed)
	rand.Seed(seed)
}

type File struct {
	Obj    interface{}
	locker sync.Locker
	cbs    chan func()
	path   string
}

func New(obj interface{}, path string, locker sync.Locker) (*File, error) {
	// check object
	if reflect.TypeOf(obj).Kind() != reflect.Ptr {
		return nil, fmt.Errorf("object must be a pointer")
	}

	// init
	file := &File{
		Obj:    obj,
		locker: locker,
		cbs:    make(chan func()),
		path:   path,
	}

	// try lock
	done := make(chan struct{})
	go func() {
		locker.Lock()
		close(done)
	}()
	select {
	case <-time.NewTimer(time.Second * 1).C:
		return nil, fmt.Errorf("lock fail")
	case <-done:
	}

	// try load from file
	dbFile, err := os.Open(path)
	if err == nil {
		defer dbFile.Close()
		err = gob.NewDecoder(dbFile).Decode(file.Obj)
		if err != nil {
			return nil, fmt.Errorf("gob file decode error: %v", err)
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
	done := make(chan struct{})
	f.cbs <- func() {
		defer close(done)
		tmpPath := f.path + "." + strconv.FormatInt(rand.Int63(), 10)
		var tmpF *os.File
		tmpF, err = os.Create(tmpPath)
		if err != nil {
			err = fmt.Errorf("open temp file error: %v", err)
			return
		}
		defer tmpF.Close()
		err = gob.NewEncoder(tmpF).Encode(f.Obj)
		if err != nil {
			return
		}
		err = os.Rename(tmpPath, f.path)
		if err != nil {
			err = fmt.Errorf("temp file rename error: %v", err)
			return
		}
	}
	<-done
	return
}

func (f *File) Close() {
	close(f.cbs)
	f.locker.Unlock()
}
