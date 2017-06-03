package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStreamingOutput(t *testing.T) {
	runner := NewRunner()
	runner.clonedir = "/"

	lines := ""
	fn := func(line string) error {
		lines += line + "\n"
		return nil
	}

	err := runner.WatchFn(fn, "docker", "version")
	assert.Nil(t, err)
	assert.NotEqual(t, "", lines)
}
