package main

import (
	"flag"
	"fmt"

	"github.com/golang/groupcache/lru"
)

type OutputHandler interface {
	AddOutput(jobid string, format string, args ...interface{})
	GetOutput(jobid string) string
}

type outputHandler struct {
	cache *lru.Cache
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
	return handler
}

func (h *outputHandler) AddOutput(jobid string, format string, args ...interface{}) {
	val, ok := h.cache.Get(jobid)
	if !ok {
		val = fmt.Sprintf(format, args...)
	} else {
		val = val.(string) + "\n" + fmt.Sprintf(format, args...)
	}
	h.cache.Add(jobid, val)
}

func (h *outputHandler) GetOutput(jobid string) string {
	val, ok := h.cache.Get(jobid)
	if !ok {
		return ""
	}
	return val.(string)
}
