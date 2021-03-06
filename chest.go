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

func (s *Chest) changed() {
	select {
	case s.sigSave <- struct{}{}:
	default:
	}
	s.dirty = true
}

func (s *Chest) Set(req *Request, response *Response) error {
	s.lock.Lock()
	s.Data[req.Key] = req.Value
	s.lock.Unlock()
	s.changed()
	return nil
}

func (s *Chest) Get(req *Request, response *Response) error {
	s.lock.RLock()
	defer s.lock.RUnlock()
	v, ok := s.Data[req.Key]
	if !ok {
		return fmt.Errorf("key not found: %s", req.Key)
	}
	response.Value = v
	return nil
}

func (s *Chest) ListAppend(req *Request, response *Response) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	v, ok := s.Data[req.Key]
	if !ok {
		value := reflect.MakeSlice(reflect.SliceOf(reflect.TypeOf((*req.Value.(*[]interface{}))[0])), 0, 0)
		for _, e := range *req.Value.(*[]interface{}) {
			value = reflect.Append(value, reflect.ValueOf(e))
		}
		s.Data[req.Key] = value.Interface()
	} else {
		value := reflect.ValueOf(v)
		if value.Type().Kind() != reflect.Slice {
			return fmt.Errorf("%s is not a slice but %T", req.Key, v)
		}
		for _, e := range *req.Value.(*[]interface{}) {
			value = reflect.Append(value, reflect.ValueOf(e))
		}
		s.Data[req.Key] = value.Interface()
	}
	s.changed()
	return nil
}

func (s *Chest) SetAdd(req *Request, response *Response) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	v, ok := s.Data[req.Key]
	if !ok {
		value := reflect.MakeMap(reflect.MapOf(reflect.TypeOf(req.Value), reflect.TypeOf((*struct{})(nil)).Elem()))
		value.SetMapIndex(reflect.ValueOf(req.Value), reflect.ValueOf(struct{}{}))
		s.Data[req.Key] = value.Interface()
	} else {
		value := reflect.ValueOf(v)
		value.SetMapIndex(reflect.ValueOf(req.Value), reflect.ValueOf(struct{}{}))
	}
	s.changed()
	return nil
}

func (s *Chest) SetExists(req *Request, response *Response) error {
	s.lock.RLock()
	defer s.lock.RUnlock()
	v, ok := s.Data[req.Key]
	if !ok {
		return fmt.Errorf("key not found: %s", req.Key)
	} else {
		if !reflect.ValueOf(v).MapIndex(reflect.ValueOf(req.Value)).IsValid() {
			return fmt.Errorf("key not exists: %s %v", req.Key, req.Value)
		}
	}
	return nil
}
