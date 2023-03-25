package pedis

import (
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

type Aof struct {
	f      *os.File
	mx     sync.Mutex
	closed bool
	rd     *Reader
	atEnd  bool
}

func (a *Aof) ReadValues(iterator func(Value) bool) error {
	a.atEnd = false
	if _, err := a.f.Seek(0, 0); err != nil {
		return err
	}

	rd := NewReader(a.f)
	for {
		v, err := rd.resp.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println("Error from readvalues: ", err)
			return err
		}

		if iterator != nil && !iterator(v) {
			return nil
		}
	}

	_, err := a.f.Seek(0, 2) // seek to the end of the file
	if err != nil {
		return err
	}

	a.atEnd = true
	return nil
}

func NewAof(path string) (*Aof, error) {
	if path == "" {
		// create aof file in the current directory
		path = "pedis.aof"
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return nil, err
	}

	aof := &Aof{
		f:  f,
		rd: NewReader(f),
	}

	// start a goroutine to flush the buffer to disk every 1 second
	// this might change in the future when we want to have more policies
	// on how to write aof to disk.
	go func() {
		for {
			aof.mx.Lock()
			if aof.closed {
				aof.mx.Unlock()
				return
			}

			aof.f.Sync()

			aof.mx.Unlock()

			time.Sleep(time.Second)
		}
	}()

	return aof, nil
}

func (a *Aof) Close() error {
	a.mx.Lock()
	defer a.mx.Unlock()

	a.closed = true

	return a.f.Close()
}

func (a *Aof) Write(b []byte) (int, error) {
	a.mx.Lock()
	defer a.mx.Unlock()

	return a.f.Write(b)
}

func (a *Aof) Append(val Value) error {
	return a.AppendMany([]Value{val})
}

func (a *Aof) AppendMany(vals []Value) error {
	var bs []byte
	for _, val := range vals {
		bytes, err := val.MarshalResp()
		if err != nil {
			return err
		}

		if bs == nil {
			bs = bytes
		} else {
			bs = append(bs, bytes...)
		}
	}

	a.mx.Lock()
	defer a.mx.Unlock()

	if a.closed {
		return errors.New("Aof file is closed")
	}

	if !a.atEnd {
		// todo: read the records to the end of the file instead of seeking here
		a.f.Seek(0, io.SeekEnd) // jump to the end of the file
		a.atEnd = true
	}

	_, err := a.f.Write(bs)
	if err != nil {
		return err
	}

	return nil
}
