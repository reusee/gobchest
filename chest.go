package gobchest

import (
	"encoding/gob"
	"fmt"
	"math/rand"
	"os"
	"reflect"
	"sync"
	"time"
)

type Chest struct {
	lock        sync.RWMutex
	Data        map[string]interface{}
	filePath    string
	handleError func(error)
	sigSave     chan struct{}
	dirty       bool
	saveTime    time.Time
}

func NewChest(filePath string) (*Chest, error) {
	info, err := os.Stat(filePath)
	var file *os.File
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
	} else if info.IsDir() {
		return nil, fmt.Errorf("%s is a directory", filePath)
	} else {
		file, err = os.Open(filePath)
		if err != nil {
			return nil, err
		}
	}
	chest := &Chest{
		Data:     make(map[string]interface{}),
		filePath: filePath,
		handleError: func(err error) {
			panic(err)
		},
		sigSave: make(chan struct{}),
	}
	// load from file
	if file != nil {
		err = gob.NewDecoder(file).Decode(chest)
		if err != nil {
			file.Close()
			return nil, err
		}
		file.Close()
	}
	return chest, nil
}

func (s *Chest) save() {
	tmpFilePath := fmt.Sprintf("%s.tmp.%d", s.filePath, rand.Uint32())
	tmpFile, err := os.Create(tmpFilePath)
	if err != nil {
		s.handleError(err)
		return
	}
	s.lock.RLock()
	err = gob.NewEncoder(tmpFile).Encode(s)
	s.lock.RUnlock()
	if err != nil {
		os.Remove(tmpFilePath)
		s.handleError(err)
		return
	}
	tmpFile.Close()
	os.Rename(tmpFilePath, s.filePath)
	s.dirty = false
	s.saveTime = time.Now()
}

func (s *Chest) Set(req *Request, response *Response) error {
	s.lock.Lock()
	s.Data[req.Key] = req.Value

	s.lock.Unlock()
	s.sigSave <- struct{}{}
	s.dirty = true
	return nil
}

func (s *Chest) Get(req *Request, response *Response) error {
	s.lock.RLock()
	v, ok := s.Data[req.Key]
	if !ok {
		s.lock.RUnlock()
		return fmt.Errorf("key not found: %s", req.Key)
	}
	response.Value = v
	s.lock.RUnlock()
	return nil
}

func (s *Chest) ListAppend(req *Request, response *Response) error {
	s.lock.Lock()
	v, ok := s.Data[req.Key]
	if !ok {
		value := reflect.MakeSlice(reflect.SliceOf(reflect.TypeOf((*req.Value.(*[]interface{}))[0])), 0, 0)
		for _, e := range *req.Value.(*[]interface{}) {
			value = reflect.Append(value, reflect.ValueOf(e))
		}
		s.Data[req.Key] = value.Interface()
	} else {
		value := reflect.ValueOf(v)
		for _, e := range *req.Value.(*[]interface{}) {
			value = reflect.Append(value, reflect.ValueOf(e))
		}
		s.Data[req.Key] = value.Interface()
	}

	s.lock.Unlock()
	s.sigSave <- struct{}{}
	s.dirty = true
	return nil
}
