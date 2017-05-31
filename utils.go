package main

import (
	"fmt"
	"sync/atomic"
)

var (
	counter uint64
)

func getCloneDir() string {
	newNumber := atomic.AddUint64(&counter, 1)
	return fmt.Sprintf("/tmp/clone%d", newNumber)
}
