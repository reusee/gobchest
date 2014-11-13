package store

import (
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
