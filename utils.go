package gobchest

import (
	"encoding/gob"
	"fmt"
	"math/rand"
	"time"
)

var (
	pt = fmt.Printf
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func Register(v interface{}) {
	gob.Register(v)
}
