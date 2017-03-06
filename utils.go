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
	newNumber = newNumber % uint64(MaxClones)
	return fmt.Sprintf("/tmp/clone%d", newNumber)
}

func getOutputFile() string {
	newNumber := atomic.AddUint64(&counter, 1)
	newNumber = newNumber % uint64(MaxOutputs)
	return fmt.Sprintf("/tmp/out%d", newNumber)
}
