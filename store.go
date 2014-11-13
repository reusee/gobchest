package store

import (
	"fmt"
	"os"
	"sync"
)

type Store struct {
	lock sync.RWMutex
	data map[string]interface{}
	file *os.File
}

func NewStore(filePath string) (*Store, error) {
	info, err := os.Stat(filePath)
	var file *os.File
	if err != nil {
		if os.IsNotExist(err) { // not exists, create one
			file, err = os.Create(filePath)
			if err != nil {
				return nil, err
			}
		} else { // open error
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
	return &Store{
		data: make(map[string]interface{}),
		file: file,
	}, nil
}

func (s *Store) Set(req *Request, response *Response) error {
	s.lock.Lock()
	s.data[req.Key] = req.Value
	s.lock.Unlock()
	return nil
}

func (s *Store) Get(req *Request, response *Response) error {
	s.lock.RLock()
	v, ok := s.data[req.Key]
	if !ok {
		return fmt.Errorf("key not found: %s", req.Key)
	}
	response.Value = v
	s.lock.RUnlock()
	return nil
}
