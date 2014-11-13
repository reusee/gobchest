package store

type RequestType uint8

var (
	Set RequestType = 1
	Get RequestType = 2
)

type Request struct {
	Type  RequestType
	Key   string
	Value interface{}
}

type Response struct {
	Value interface{}
}
