package gobchest

import "encoding/gob"

func init() {
	gob.Register(new([]interface{}))
}

type RequestType uint8

var (
	Set        RequestType = 1
	Get        RequestType = 2
	ListAppend RequestType = 3
)

type Request struct {
	Type  RequestType
	Key   string
	Value interface{}
}

type Response struct {
	Value interface{}
}
