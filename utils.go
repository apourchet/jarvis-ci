package main

import (
	"fmt"
	"io/ioutil"
	"sync/atomic"

	"github.com/golang/glog"
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

func writeOutput(out []byte, outputFile string) {
	glog.Infof("Writing output to %s: %s", outputFile, string(out))
	err := ioutil.WriteFile(outputFile, out, 0644)
	if err != nil {
		glog.Warningf("Failed to write output to %s", outputFile)
	}
}
