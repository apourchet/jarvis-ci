package main

import (
	"flag"
	"fmt"
	"sync"

	"github.com/golang/groupcache/lru"
)

type OutputHandler interface {
	AddOutput(jobid string, format string, args ...interface{})
	GetOutput(jobid string) string
}

type outputHandler struct {
	cache *lru.Cache
	lock  *sync.Mutex
}

var (
	OutputCacheSize int
)

var _ OutputHandler = &outputHandler{}

func init() {
	flag.IntVar(&OutputCacheSize, "output-cache", 20, "Number of build outputs to keep in LRU cache")
}

func DefaultOutputHandler() *outputHandler {
	return NewOutputHandler(OutputCacheSize)
}

func NewOutputHandler(size int) *outputHandler {
	handler := &outputHandler{}
	handler.cache = lru.New(size)
	handler.lock = &sync.Mutex{}
	return handler
}

func (h *outputHandler) AddOutput(jobid string, format string, args ...interface{}) {
	h.lock.Lock()
	defer h.lock.Unlock()
	val, ok := h.cache.Get(jobid)
	if !ok {
		val = fmt.Sprintf(format, args...)
	} else {
		val = val.(string) + "\n" + fmt.Sprintf(format, args...)
	}
	h.cache.Add(jobid, val)
}

func (h *outputHandler) GetOutput(jobid string) string {
	h.lock.Lock()
	defer h.lock.Unlock()
	val, ok := h.cache.Get(jobid)
	if !ok {
		return ""
	}
	return val.(string)
}
