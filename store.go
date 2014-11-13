package store

import (
	"encoding/gob"
	"fmt"
	"math/rand"
	"os"
	"sync"
)

type Store struct {
	lock        sync.RWMutex
	Data        map[string]interface{}
	filePath    string
	handleError func(error)
}

func NewStore(filePath string) (*Store, error) {
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
	store := &Store{
		Data:     make(map[string]interface{}),
		filePath: filePath,
		handleError: func(err error) {
			panic(err)
		},
	}
	// load from file
	if file != nil {
		err = gob.NewDecoder(file).Decode(store)
		if err != nil {
			file.Close()
			return nil, err
		}
		file.Close()
	}
	return store, nil
}

func (s *Store) save() {
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
}

func (s *Store) Set(req *Request, response *Response) error {
	s.lock.Lock()
	s.Data[req.Key] = req.Value
	s.lock.Unlock()
	return nil
}

func (s *Store) Get(req *Request, response *Response) error {
	s.lock.RLock()
	v, ok := s.Data[req.Key]
	if !ok {
		return fmt.Errorf("key not found: %s", req.Key)
	}
	response.Value = v
	s.lock.RUnlock()
	return nil
}
