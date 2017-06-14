package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"sync"

	"github.com/golang/glog"
)

type Runner struct {
	clonedir string
}

func NewRunner() Runner {
	runner := Runner{}
	runner.clonedir = getCloneDir()
	return runner
}

func (r Runner) CloneRepo(cloneURL string, ref string) error {
	glog.Infof("Cloning from %s into %s", ref, r.clonedir)
	err := exec.Command("git", "clone", cloneURL, r.clonedir, "--depth", "1").Run()
	if err != nil {
		return fmt.Errorf("Failed to clone directory %s into %s: %v", cloneURL, r.clonedir, err)
	}

	// Fetch the ref
	cmd := exec.Command("git", "fetch", "origin", ref, "--depth", "1")
	cmd.Dir = r.clonedir
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("Failed to fetch ref %s: %v", ref, err)
	}

	// Checkout the fetched head
	cmd = exec.Command("git", "checkout", "FETCH_HEAD")
	cmd.Dir = r.clonedir
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("Failed to checkout FETCH_HEAD: %v", err)
	}
	return nil
}

func (r Runner) Checkout(head string) error {
	glog.Infof("Checking out head %s", head)
	cmd := exec.Command("git", "checkout", head)
	cmd.Dir = r.clonedir
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("Failed to checkout head %s in %s: %v", head, r.clonedir, err)
	}
	return nil
}

func (r Runner) Run(program string, args ...string) ([]byte, error) {
	glog.Infof("Running `%s %v`", program, args)
	cmd := exec.Command(program, args...)
	cmd.Dir = r.clonedir
	return cmd.CombinedOutput()
}

type item struct {
	output string
	err    error
}

func (r Runner) Watch(program string, args ...string) (chan item, error) {
	glog.Infof("Watching `%s %v`", program, args)

	out := make(chan item, 0)
	cmd := exec.Command(program, args...)
	cmd.Dir = r.clonedir

	cmdReader, err := cmd.StdoutPipe()
	if err != nil {
		close(out)
		return out, err
	}

	err = cmd.Start()
	if err != nil {
		close(out)
		return out, err
	}

	wg := &sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(cmdReader)
		for scanner.Scan() {
			if err := scanner.Err(); err != nil {
				out <- item{"", err}
				return
			}
			out <- item{scanner.Text(), nil}
		}
	}()

	go func() {
		defer wg.Done()
		err = cmd.Wait()
		if err != nil {
			out <- item{"", err}
		}
	}()

	go func() {
		wg.Wait()
		close(out)
	}()

	return out, nil
}

func (r Runner) WatchFn(fn func(string) error, program string, args ...string) error {
	items, err := r.Watch(program, args...)
	if err != nil {
		return err
	}

	for item := range items {
		if item.err != nil {
			return item.err
		}

		err = fn(item.output)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r Runner) Cleanup() {
	err := os.RemoveAll(r.clonedir)
	if err != nil {
		glog.Errorf("Failed to cleanup runner %v: %v", r, err)
	}
}
